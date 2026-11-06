package main

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	key := "TEST_ROCK_SQL_ENV"
	_ = os.Unsetenv(key)

	if got := getEnv(key, "default_val"); got != "default_val" {
		t.Errorf("expected fallback 'default_val', got %q", got)
	}

	_ = os.Setenv(key, "custom_val")
	defer os.Unsetenv(key)

	if got := getEnv(key, "default_val"); got != "custom_val" {
		t.Errorf("expected 'custom_val', got %q", got)
	}
}

func TestIdentRe(t *testing.T) {
	valid := []string{"users", "periodic_table", "table_123", "public", "SCHEMA"}
	for _, id := range valid {
		if !identRe.MatchString(id) {
			t.Errorf("expected %q to be valid identifier", id)
		}
	}

	invalid := []string{
		"users; DROP TABLE users;--",
		"table name",
		"table\"name",
		"users' OR '1'='1",
		"table$1",
		"",
	}
	for _, id := range invalid {
		if identRe.MatchString(id) {
			t.Errorf("expected %q to be rejected as invalid identifier", id)
		}
	}
}

func TestDatabaseDefinitions(t *testing.T) {
	seen := make(map[string]bool)
	for _, db := range databases {
		if db.Name == "" {
			t.Error("database name cannot be empty")
		}
		if db.Blurb == "" {
			t.Errorf("database %q blurb cannot be empty", db.Name)
		}
		if seen[db.Name] {
			t.Errorf("duplicate database name %q", db.Name)
		}
		seen[db.Name] = true
	}
}

func TestPoolUnknownDatabase(t *testing.T) {
	_, err := pool("nonexistent_db_12345")
	if err == nil {
		t.Fatal("expected error for unknown database, got nil")
	}
}
