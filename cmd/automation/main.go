// Command automation is TrackLin.es's nightly clean-up (TL-431): the purge (tickets 30 days in the bin,
// organizations deleted 30 days ago, year-old attachments, expired Chat link codes) and then deleting
// files from UploadThing whose 30 days are up. It runs both and exits: a Railway cron service, far
// lighter than starting the whole backend for it. The rules are the backend's own (package retention),
// so they live in one place; it runs no migrations, since the tables are the backend's.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	bugfixes "github.com/bugfixes/go-bugfixes"
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"github.com/tracklines/backend/retention"
)

func main() {
	if err := run(context.Background()); err != nil {
		_ = logs.Errorf("automation: %v", err) // to Bugfixes, as the service "automation"
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// the backend's own config library, so the same variables work here: RDS_URL (or RDS_HOSTNAME,
	// RDS_PORT, RDS_DB, RDS_USERNAME, RDS_PASSWORD), BUGFIXES_AGENT_KEY/SECRET, BUGFIXES_LOCAL_ONLY
	c := ConfigBuilder.NewConfigNoVault()
	if err := c.Build(ConfigBuilder.Local, ConfigBuilder.Postgres, ConfigBuilder.Bugfixes); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	bugfixes.SetDefaultConfig(bugfixes.Config{
		Server:      c.Bugfixes.Server,
		AgentKey:    c.Bugfixes.AgentKey,
		AgentSecret: c.Bugfixes.AgentSecret,
		Service:     "automation",
		LocalOnly:   c.Local.KeepLocal || c.Bugfixes.AgentKey == "",
	})
	db, err := c.Database.GetPGXPoolClient(ctx)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	files := retention.UploadThing{Token: os.Getenv("UPLOADTHING_TOKEN")}
	return nightly(ctx,
		func(ctx context.Context) (string, error) {
			r, err := retention.Purge(ctx, db, retention.ToInbox(db), time.Now())
			return r.String(), err
		},
		func(ctx context.Context) (string, error) {
			n, err := retention.Sweeper{DB: db, Files: files}.SweepAll(ctx)
			return fmt.Sprintf("deleted %d file(s) from UploadThing", n), err
		})
}

// job is one of the night's jobs: what it did, for the log.
type job func(ctx context.Context) (string, error)

// nightly runs the purge, then the file sweep. They're independent: one failing doesn't stop the
// other, and both failures are returned.
func nightly(ctx context.Context, purge, sweep job) error {
	var failed []error
	for _, j := range []struct {
		name string
		run  job
	}{{"purge", purge}, {"sweep-files", sweep}} {
		done, err := j.run(ctx)
		logs.Logf("%s: %s", j.name, done)
		if err != nil {
			failed = append(failed, fmt.Errorf("%s: %w", j.name, err))
		}
	}
	return errors.Join(failed...)
}
