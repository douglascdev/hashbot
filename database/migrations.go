package database

import (
	"database/sql"
	"fmt"
	"sort"
)

type DBMigration struct {
	Version int
	Stmts   []string
}

// makes migrations sortable by version(implements sort.Interface)
type DBMigrations struct {
	Migrations []DBMigration
}

func (m *DBMigrations) Len() int {
	return len(m.Migrations)
}

func (m *DBMigrations) Swap(i, j int) {
	m.Migrations[i], m.Migrations[j] = m.Migrations[j], m.Migrations[i]
}

func (m *DBMigrations) Less(i, j int) bool {
	return m.Migrations[i].Version < m.Migrations[j].Version
}

var CurrentSchemaStmts = []string{
	`CREATE TABLE user (
                       id TEXT NOT NULL PRIMARY KEY,
                       name TEXT NOT NULL,
                       permission_id INTEGER NOT NULL,
                       bot_is_joined BOOL NOT NULL DEFAULT false,
                       language TEXT NOT NULL DEFAULT english,
                       FOREIGN KEY (permission_id) REFERENCES permission(id)
               )`,
	`CREATE TABLE user_editor (
						id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                        user_id TEXT NOT NULL,
                        editor_id TEXT NOT NULL,
                        FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
                        FOREIGN KEY (editor_id) REFERENCES user(id) ON DELETE CASCADE
	)`,
	`CREATE TABLE permission (
                       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                       name TEXT NOT NULL,
                       is_ignored BOOL NOT NULL DEFAULT false,
                       is_bot_admin BOOL NOT NULL DEFAULT false
               )`,
	`CREATE TABLE command (
                       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                       name TEXT NOT NULL
               )`,
	`CREATE INDEX idx_name ON command(name)`,
	`CREATE TABLE user_command (
                       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                       user_id TEXT NOT NULL,
                       command_id INTEGER NOT NULL,
                       is_enabled BOOL NOT NULL DEFAULT true,
                       last_used INTEGER NOT NULL DEFAULT 1726849749,
                       FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
                       FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
               )`,
	`CREATE INDEX idx_is_enabled ON user_command(is_enabled)`,
	`CREATE INDEX idx_user_command ON user_command(user_id, command_id)`,
	`CREATE TABLE user_command_data (
                       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                       user_id INTEGER NOT NULL,
                       command_id INTEGER NOT NULL,
                       last_used INTEGER NOT NULL DEFAULT 1726849749,
                       opted_out BOOL NOT NULL DEFAULT false,
                       FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
                       FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
               )`,
	`CREATE INDEX idx_user_command_data ON user_command_data(user_id, command_id, last_used)`,
	`CREATE TABLE rpg_item (
                       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                       name TEXT NOT NULL,
                       description TEXT NOT NULL
               )`,
	`CREATE TABLE rpg_user_item (
                       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                       user_id TEXT NOT NULL,
                       rpg_item_id INTEGER NOT NULL,
                       amount INTEGER NOT NULL,

                       FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
                       FOREIGN KEY (rpg_item_id) REFERENCES rpg_item(id) ON DELETE CASCADE
               )`,
	`CREATE TABLE app_data (
					   id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					   migration_version INTEGER NOT NULL
    )`,

	// DML
	`INSERT INTO app_data (migration_version) VALUES (1)`,

	`INSERT INTO permission (name) VALUES ('user')`,
	`INSERT INTO permission (name, is_ignored) VALUES ('banned', true)`,
	`INSERT INTO permission (name, is_bot_admin) VALUES ('admin', true)`,

	`INSERT INTO rpg_item (name, description) VALUES ('buttinho', 'The most widely used currency in the seven seas.')`,
}

var Migrations = DBMigrations{
	Migrations: []DBMigration{
		{Version: 1, Stmts: []string{
			"INSERT INTO command (name) VALUES ('butt')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'butt'
				), true FROM user
			`,
		}},
		{Version: 2, Stmts: []string{
			"INSERT INTO command (name) VALUES ('help')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'help'
				), true FROM user
			`,
		}},
		{Version: 3, Stmts: []string{
			"INSERT INTO command (name) VALUES ('explore')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'explore'
				), true FROM user`,
			`CREATE TABLE rpg_item (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT NOT NULL
			)`,
			`CREATE TABLE rpg_user_item (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				rpg_item_id INTEGER NOT NULL,
				amount INTEGER NOT NULL,

				FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
				FOREIGN KEY (rpg_item_id) REFERENCES rpg_item(id) ON DELETE CASCADE
			)`,
			`INSERT INTO rpg_item (name, description) VALUES ('buttinho', 'The most widely used currency in the seven seas.')`,
		}},
		{Version: 4, Stmts: []string{
			"INSERT INTO command (name) VALUES ('enable'), ('disable')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'enable'
				), true FROM user`,
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'disable'
				), true FROM user`,
		}},
		{Version: 5, Stmts: []string{
			"ALTER TABLE user_command ADD last_used INTEGER NOT NULL DEFAULT 1726849749",
		}},
		{Version: 6, Stmts: []string{
			`CREATE INDEX idx_user_command ON user_command(user_id, command_id)`,
			`CREATE TABLE user_command_cooldown (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				command_id INTEGER NOT NULL,
				last_used INTEGER NOT NULL DEFAULT 1726849749,
				FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
				FOREIGN KEY (command_id) REFERENCES command(id) ON DELETE CASCADE
			)`,

			// DML
			`INSERT INTO permission (name) VALUES ('user')`,
			`INSERT INTO permission (name, is_ignored) VALUES ('banned', true)`,
			`INSERT INTO permission (name, is_bot_admin) VALUES ('admin', true)`,

			`INSERT INTO rpg_item (name, description) VALUES ('buttinho', 'The most widely used currency in the seven seas.')`,
		}},
		{Version: 7, Stmts: []string{
			"ALTER TABLE user_command_cooldown RENAME TO user_command_data",
			`
			CREATE TABLE app_data (
					   id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					   migration_version INTEGER NOT NULL
			)`,
			`INSERT INTO app_data (migration_version) VALUES (7)`,
		}},
		{Version: 8, Stmts: []string{
			`CREATE TABLE user_editor (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				editor_id TEXT NOT NULL,
				FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
				FOREIGN KEY (editor_id) REFERENCES user(id) ON DELETE CASCADE
			)`,

			"INSERT INTO command (name) VALUES ('editor'), ('remove'), ('yoink')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'editor'
				), true FROM user`,
			`INSERT INTO user_command_data (user_id, command_id)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'editor'
				) FROM user`,

			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'remove'
				), true FROM user`,
			`INSERT INTO user_command_data (user_id, command_id)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'remove'
				) FROM user`,

			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'yoink'
				), true FROM user`,
			`INSERT INTO user_command_data (user_id, command_id)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'yoink'
				) FROM user`,
		}},
		{Version: 9, Stmts: []string{
			"INSERT INTO command (name) VALUES ('add')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'add'
				), true FROM user`,
			`INSERT INTO user_command_data (user_id, command_id)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'add'
				) FROM user`,
		}},
		{Version: 10, Stmts: []string{
			"INSERT INTO command (name) VALUES ('yoink')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'yoink'
				), true FROM user`,
			`INSERT INTO user_command_data (user_id, command_id)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'yoink'
				) FROM user`,
		}},
		{Version: 11, Stmts: []string{
			"ALTER TABLE user ADD COLUMN language TEXT NOT NULL DEFAULT en",
			"INSERT INTO command (name) VALUES ('language')",
			`INSERT INTO user_command (user_id, command_id, is_enabled)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'language'
				), true FROM user`,
			`INSERT INTO user_command_data (user_id, command_id)
				SELECT id, (
					SELECT c.id FROM command c WHERE c.name = 'language'
				) FROM user`,
		}},
	}}

func RunMigrations(tx *sql.Tx, migrations *DBMigrations, currentSchemaStmts []string) error {
	// sort migrations by version
	sort.Sort(migrations)

	var err error

	// migrations to be applied sequentially until the currentVersion
	currentVersion, err := SelectMigrationVersion(tx)
	if err != nil {
		// no version found, create schema from scratch
		for _, stmt := range CurrentSchemaStmts {
			_, err := tx.Exec(stmt)
			if err != nil {
				return fmt.Errorf("failed to execute current schema on statement '%s' with error: %w", stmt, err)
			}
		}
	} else {
		// migrate current schema
		for i := currentVersion; i < len(migrations.Migrations); i++ {
			migration := migrations.Migrations[i]

			for _, stmt := range migration.Stmts {
				_, err = tx.Exec(stmt)
				if err != nil {
					return fmt.Errorf("failed to execute migration for statement '%s' with error: %w", stmt, err)
				}
			}
		}
	}

	currentVersion = migrations.Migrations[len(migrations.Migrations)-1].Version
	err = UpdateMigrationVersion(tx, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to update migration version: %w", err)
	}
	return nil
}
