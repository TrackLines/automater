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
