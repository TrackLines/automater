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
	"net/url"
	"os"
	"time"

	bugfixes "github.com/bugfixes/go-bugfixes"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tracklines/backend/retention"
)

func main() {
	if err := run(context.Background()); err != nil {
		_ = logs.Errorf("automation: %v", err) // to Bugfixes, as the service "automation"
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	setupBugfixes()
	dsn, err := databaseURL(os.Getenv)
	if err != nil {
		return err
	}
	db, err := pgxpool.New(ctx, dsn)
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

// databaseURL is DATABASE_URL (Railway's Postgres provides it), or the backend's own RDS_* variables,
// so either can be copied across.
func databaseURL(env func(string) string) (string, error) {
	if u := env("DATABASE_URL"); u != "" {
		return u, nil
	}
	host, db := env("RDS_HOSTNAME"), env("RDS_DB")
	if host == "" || db == "" {
		return "", errors.New("no database: set DATABASE_URL, or RDS_HOSTNAME, RDS_PORT, RDS_DB, RDS_USERNAME and RDS_PASSWORD")
	}
	port := env("RDS_PORT")
	if port == "" {
		port = "5432"
	}
	u := url.URL{Scheme: "postgres", User: url.UserPassword(env("RDS_USERNAME"), env("RDS_PASSWORD")), Host: host + ":" + port, Path: "/" + db}
	return u.String(), nil
}

// setupBugfixes reports failures as the service "automation", with BUGFIXES_AGENT_KEY and
// BUGFIXES_AGENT_SECRET (the agent made for this service). Without them, failures are only logged.
func setupBugfixes() {
	key, secret := os.Getenv("BUGFIXES_AGENT_KEY"), os.Getenv("BUGFIXES_AGENT_SECRET")
	bugfixes.SetDefaultConfig(bugfixes.Config{
		Server:      bugfixes.DefaultServer,
		AgentKey:    key,
		AgentSecret: secret,
		Service:     "automation",
		LocalOnly:   key == "" || secret == "",
	})
}
