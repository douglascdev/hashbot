package database

import (
	"bytes"
	"database/sql"
	"hashbot/config"
	"testing"
)

var testDB *sql.DB

func init() {
	var err error
	testDB, err = generateTestDB()
	if err != nil {
		panic(err)
	}
}

func generateTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	// pragmas that should speed up sqlite for testing
	db.Exec("PRAGMA synchronous = OFF;")
	db.Exec("PRAGMA journal_mode = MEMORY;")
	db.Exec("PRAGMA temp_store = MEMORY;")
	return db, nil
}

func generateTestConfig() (*config.Config, error) {
	template, err := config.ConfigTemplateJSON()
	if err != nil {
		return nil, err
	}
	var cfg *config.Config
	cfg, err = config.LoadConfig(template)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestGenerateTestDB(t *testing.T) {
	db, err := generateTestDB()
	if err != nil {
		t.Errorf("failed to init test database: %v", err)
	}
	defer db.Close()
}

func TestInitDB(t *testing.T) {
	cfg, err := generateTestConfig()
	if err != nil {
		t.Errorf("failed to generate test config: %v", err)
	}

	var (
		reader = new(bytes.Buffer)
		data   []byte
	)

	data, err = config.MarshalConfig(cfg)
	if err != nil {
		t.Errorf("failed to marshal test config: %v", err)
	}
	reader.Write(data)

	db, err := InitDB("sqlite3", "file:data.db?mode=memory", reader)
	if err != nil {
		t.Errorf("failed to run InitDB: %v", err)
	}
	defer db.Close()

	cfg, err = config.LoadConfig(data)
	if err != nil {
		t.Errorf("failed to load written config: %v", err)
	}
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
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

func TestInsertCommands(t *testing.T) {
	tx, err := testDB.Begin()
	defer tx.Rollback()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}
	err = RunMigrations(tx, &Migrations, CurrentSchemaStmts)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	err = InsertCommands(tx, "test")
	if err != nil {
		t.Errorf("failed to insert commands: %v", err)
	}

	res := tx.QueryRow("SELECT name FROM command")
	var name string
	err = res.Scan(&name)
	if err != nil {
		t.Errorf("failed to scan name value: %v", err)
	}
	if name != "test" {
		t.Errorf("unexpected name value: %s", name)
	}
}

func TestInsertUsers(t *testing.T) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	err = RunMigrations(tx, &Migrations, CurrentSchemaStmts)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	err = InsertUsers(
		tx,
		false,
		"en",
		struct {
			ID   string
			Name string
		}{
			ID:   "test",
			Name: "test",
		},
	)
	if err != nil {
		t.Errorf("failed to insert users: %v", err)
	}
}

func TestInsertUserCommands(t *testing.T) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	err = RunMigrations(tx, &Migrations, CurrentSchemaStmts)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	err = InsertUsers(tx, true, "en", struct {
		ID   string
		Name string
	}{"test", "test"})
	if err != nil {
		t.Errorf("failed to insert users: %v", err)
	}

	err = InsertCommands(tx, "test")
	if err != nil {
		t.Errorf("failed to insert commands: %v", err)
	}

	err = InsertUserCommands(tx, "test", "test")
	if err != nil {
		t.Errorf("failed to insert user commands: %v", err)
	}
}

func TestUpdateUserPermission(t *testing.T) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	err = RunMigrations(tx, &Migrations, CurrentSchemaStmts)
	if err != nil {
		t.Errorf("failed to run migrations: %v", err)
	}

	err = InsertUsers(tx, false, "en", struct {
		ID   string
		Name string
	}{"test", "test"})
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	var username string
	err = tx.QueryRow("SELECT id FROM user WHERE name = 'test'").Scan(&username)
	if err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}

	err = UpdateUserPermission(tx, username, "admin")
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	var permissionName string
	err = tx.QueryRow("SELECT p.name FROM permission p INNER JOIN user u ON u.permission_id = p.id WHERE u.name = 'test'").Scan(&permissionName)
	if err != nil {
		t.Fatalf("failed to get permission name: %v", err)
	}
	if permissionName != "admin" {
		t.Fatalf("permission name was not updated: %s", permissionName)
	}
}

func TestSelectIsUserIgnored(t *testing.T) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	err = RunMigrations(tx, &Migrations, CurrentSchemaStmts)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	var bannedPermissionID int
	err = tx.QueryRow("SELECT id FROM permission WHERE name = 'banned'").Scan(&bannedPermissionID)
	if err != nil {
		t.Fatalf("failed to get banned permission id: %v", err)
	}

	var adminPermissionID int
	err = tx.QueryRow("SELECT id FROM permission WHERE name = 'admin'").Scan(&adminPermissionID)
	if err != nil {
		t.Fatalf("failed to get admin permission id: %v", err)
	}

	users := []struct {
		ID   string
		Name string
	}{
		{"test1", "test1"},
		{"test2", "test2"},
	}

	err = InsertUsers(tx, false, "en", users...)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	var test1ID, test2ID string
	err = tx.QueryRow("SELECT id FROM user WHERE name = 'test1'").Scan(&test1ID)
	if err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}

	err = tx.QueryRow("SELECT id FROM user WHERE name = 'test2'").Scan(&test2ID)
	if err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}

	err = UpdateUserPermission(tx, users[0].Name, "banned")
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	err = UpdateUserPermission(tx, users[1].Name, "admin")
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	// ensure banned users are ignored
	var isIgnored bool
	isIgnored, err = SelectIsUserIgnored(tx, test1ID)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	if !isIgnored {
		t.Fatal("expected banned user to be ignored")
	}

	// ensure admin users aren't ignored
	isIgnored, err = SelectIsUserIgnored(tx, test2ID)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	if isIgnored {
		t.Fatal("expected admin user to not be ignored")
	}
}
