// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strings"

	"github.com/xwb1989/sqlparser"
)

type WriteQueryError struct {
	Query string
}

func (e *WriteQueryError) Error() string {
	return ErrorPrefix + "Write operations are not allowed in read-only mode. Query attempted: " + truncateQuery(e.Query)
}

func (e *WriteQueryError) IsSqlcmdErr() bool {
	return true
}

func truncateQuery(q string) string {
	const maxLen = 100
	if len(q) > maxLen {
		return q[:maxLen] + "..."
	}
	return q
}

func IsReadOnlyQuery(query string) bool {
	if strings.TrimSpace(query) == "" {
		return true
	}

	queries := splitStatements(query)
	for _, q := range queries {
		if !isQueryReadOnly(q) {
			return false
		}
	}
	return true
}

func splitStatements(query string) []string {
	var queries []string
	stmt, remains, err := sqlparser.SplitStatement(query)
	for err == nil && stmt != "" {
		queries = append(queries, strings.TrimSpace(stmt))
		stmt, remains, err = sqlparser.SplitStatement(remains)
	}
	if remains != "" && strings.TrimSpace(remains) != "" {
		queries = append(queries, strings.TrimSpace(remains))
	}
	return queries
}

func isQueryReadOnly(query string) bool {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return true
	}

	typ := sqlparser.Preview(trimmed)

	switch typ {
	case sqlparser.StmtSelect,
		sqlparser.StmtShow,
		sqlparser.StmtUse,
		sqlparser.StmtSet,
		sqlparser.StmtBegin,
		sqlparser.StmtCommit,
		sqlparser.StmtRollback,
		sqlparser.StmtOther,
		sqlparser.StmtUnknown,
		sqlparser.StmtComment:
		return true
	default:
		return false
	}
}
