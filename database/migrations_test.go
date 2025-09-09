package database

import (
	"testing"
)

func TestRunCurrentMigrations(t *testing.T) {
	tx, err := testDB.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}

	err = RunMigrations(tx, &Migrations, CurrentSchemaStmts)
	if err != nil {
		t.Errorf("failed to run migrations with current schema: %v", err)
	}
	version, err := SelectMigrationVersion(tx)
	if err != nil {
		t.Errorf("failed to retrieve migration version: %v", err)
	}
	expectedVersion := Migrations.Migrations[len(Migrations.Migrations)-1].Version
	if version != expectedVersion {
		t.Errorf("migration failed to update database version, expected %d, got %d", expectedVersion, version)
	}
}

func TestRunMigrations(t *testing.T) {
	migrations := DBMigrations{
		Migrations: []DBMigration{
			{Version: 1, Stmts: []string{
				"INSERT INTO test (name) VALUES ('test')",
			}},
			{Version: 2, Stmts: []string{
				"INSERT INTO test (name) VALUES ('test')",
			}},
		},
	}

	var (
		err error
	)

	tx, err := testDB.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}

	currentSchemaStmts := []string{
		"CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT NOT NULL)",
		`CREATE TABLE app_data (
					   id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					   migration_version INTEGER NOT NULL
		)`,
		`INSERT INTO app_data (migration_version) VALUES (0)`,
	}
	for _, stmt := range currentSchemaStmts {
		_, err = tx.Exec(stmt)
		if err != nil {
			t.Error(err)
		}
	}
	err = RunMigrations(tx, &migrations, []string{})
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	migrationsVersion, err := SelectMigrationVersion(tx)
	if err != nil {
		t.Errorf("failed to get migrations version: %v", err)
	}
	if migrationsVersion != 2 {
		t.Errorf("expected version 2, got %d", migrationsVersion)
	}

	res := tx.QueryRow("SELECT id, name FROM test")
	var (
		id   int
		name string
	)
	err = res.Scan(&id, &name)
	if err != nil {
		t.Errorf("failed to scan name value: %v", err)
	}
	if name != "test" {
		t.Errorf("unexpected name value: %s", name)
	}
}
