package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Both jobs run every night, whichever fails; both failures come back.
func TestNightly(t *testing.T) {
	var ran []string
	ok := func(name string) job {
		return func(context.Context) (string, error) { ran = append(ran, name); return "fine", nil }
	}
	broken := func(name string) job {
		return func(context.Context) (string, error) { ran = append(ran, name); return "", errors.New(name + " broke") }
	}
	if err := nightly(context.Background(), ok("purge"), ok("sweep")); err != nil || strings.Join(ran, ",") != "purge,sweep" {
		t.Fatalf("both fine: %v %v", ran, err)
	}
	ran = nil
	err := nightly(context.Background(), broken("purge"), ok("sweep"))
	if strings.Join(ran, ",") != "purge,sweep" || err == nil || !strings.Contains(err.Error(), "purge: purge broke") {
		t.Fatalf("a failed purge stopped the sweep: %v %v", ran, err)
	}
	ran = nil
	err = nightly(context.Background(), broken("purge"), broken("sweep"))
	if err == nil || !strings.Contains(err.Error(), "purge broke") || !strings.Contains(err.Error(), "sweep broke") {
		t.Fatalf("both failures: %v", err)
	}
}

func TestDatabaseURL(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if u, _ := databaseURL(env(map[string]string{"DATABASE_URL": "postgres://a@b/c", "RDS_HOSTNAME": "ignored"})); u != "postgres://a@b/c" {
		t.Fatalf("DATABASE_URL first: %q", u)
	}
	u, err := databaseURL(env(map[string]string{"RDS_HOSTNAME": "db.internal", "RDS_DB": "tracklines", "RDS_USERNAME": "tl", "RDS_PASSWORD": "p@ss/word"}))
	if err != nil || u != "postgres://tl:p%40ss%2Fword@db.internal:5432/tracklines" {
		t.Fatalf("from RDS_*: %q %v", u, err)
	}
	if _, err := databaseURL(env(nil)); err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("no database: %v", err)
	}
}
