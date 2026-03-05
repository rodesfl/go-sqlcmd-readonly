// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/sql"
)

// Query defines the `sqlcmd query` command
type Query struct {
	cmdparser.Cmd

	text         string
	database     string
	maxRows      int
	outputFormat string
}

func (c *Query) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "query",
		Short: localizer.Sprintf("Run a query against the current context"),
		Examples: []cmdparser.ExampleOptions{
			{Description: localizer.Sprintf("Run a query"), Steps: []string{
				`sqlcmd query "SELECT @@SERVERNAME"`,
				`sqlcmd query --text "SELECT @@SERVERNAME"`,
				`sqlcmd query --query "SELECT @@SERVERNAME"`,
			}},
			{Description: localizer.Sprintf("Run a query using [%s] database", "master"), Steps: []string{
				`sqlcmd query "SELECT DB_NAME()" --database master`,
			}},
			{Description: localizer.Sprintf("Limit query results to 100 rows"), Steps: []string{
				`sqlcmd query "SELECT * FROM sys.objects" --max-rows 100`,
			}},
			{Description: localizer.Sprintf("Set new default database"), Steps: []string{
				fmt.Sprintf(`sqlcmd query "ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [tempdb]" --database master`,
					pal.UserName()),
			}},
		},
		Run: c.run,
		FirstArgAlternativeForFlag: &cmdparser.AlternativeForFlagOptions{
			Flag:  "text",
			Value: &c.text,
		},
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.text,
		Name:      "text",
		Shorthand: "t",
		Usage:     localizer.Sprintf("Command text to run")})

	// BUG(stuartpa): Decide on if --text or --query is best (or leave both for convenience)
	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.text,
		Name:      "query",
		Shorthand: "q",
		Usage:     localizer.Sprintf("Command text to run")})

	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.database,
		Name:      "database",
		Shorthand: "d",
		Usage:     localizer.Sprintf("Database to use")})

	c.AddFlag(cmdparser.FlagOptions{
		Int:        &c.maxRows,
		DefaultInt: 500,
		Name:       "max-rows",
		Shorthand:  "m",
		Usage:      localizer.Sprintf("Maximum number of rows to return")})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.outputFormat,
		Name:   "output",
		Usage:  localizer.Sprintf("Output format: table or ndjson")})
}

// run executes the Query command.
// It connects to a SQL Server endpoint using the current context from the config file,
// and either runs an interactive SQL console or executes the provided query.
// If an error occurs, it is handled by the CheckErr function.
func (c *Query) run() {
	endpoint, user := config.CurrentContext()

	s := sql.New(sql.SqlOptions{})
	if c.text == "" {
		s.Connect(endpoint, user, sql.ConnectOptions{
			Database:     c.database,
			Interactive:  true,
			OutputFormat: "table",
		})
	} else {
		s.Connect(endpoint, user, sql.ConnectOptions{
			Database:     c.database,
			Interactive:  false,
			MaxRows:      c.maxRows,
			OutputFormat: c.outputFormat,
		})
	}

	s.Query(c.text)
}
