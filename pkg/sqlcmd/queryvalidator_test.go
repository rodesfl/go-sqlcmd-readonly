// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsReadOnlyQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"Empty query", "", true},
		{"Whitespace only", "   ", true},
		{"Simple SELECT", "SELECT * FROM users", true},
		{"SELECT with WHERE", "SELECT id, name FROM users WHERE id = 1", true},
		{"SELECT with JOIN", "SELECT * FROM a JOIN b ON a.id = b.id", true},
		{"SELECT with subquery", "SELECT * FROM users WHERE id IN (SELECT id FROM active_users)", true},
		{"INSERT", "INSERT INTO users (name) VALUES ('test')", false},
		{"UPDATE", "UPDATE users SET name = 'test' WHERE id = 1", false},
		{"DELETE", "DELETE FROM users WHERE id = 1", false},
		{"DROP TABLE", "DROP TABLE users", false},
		{"CREATE TABLE", "CREATE TABLE users (id INT)", false},
		{"ALTER TABLE", "ALTER TABLE users ADD name VARCHAR(100)", false},
		{"TRUNCATE", "TRUNCATE TABLE users", false},
		{"Multiple SELECTs", "SELECT 1; SELECT 2", true},
		{"SELECT then INSERT", "SELECT 1; INSERT INTO t VALUES(1)", false},
		{"REPLACE", "REPLACE INTO users (id, name) VALUES (1, 'test')", false},
		{"USE database", "USE mydb", true},
		{"SET variable", "SET @var = 1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReadOnlyQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteQueryError(t *testing.T) {
	err := &WriteQueryError{Query: "INSERT INTO users VALUES (1)"}
	assert.True(t, err.IsSqlcmdErr())
	assert.Contains(t, err.Error(), "Write operations are not allowed")
	assert.Contains(t, err.Error(), "INSERT")
}
