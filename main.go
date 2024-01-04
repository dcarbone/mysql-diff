package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli/v2"
)

const (
	flagConn   = "conn"
	flagPretty = "pretty"
)

func preRun(cctx *cli.Context) (mysqlConns, error) {
	connConfigs, err := parseConnFlags(cctx)
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	conns, err := openConnections(connConfigs)
	if err != nil {
		return nil, fmt.Errorf("error opening connections: %w", err)
	}

	return conns, nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     flagConn,
				Aliases:  []string{"C"},
				Usage:    "A single MySQL connection with structure: \"addr=$addr user=$user pass=$pass db=$db[ db=$dbX]\"",
				Required: true,
			},
		},
		Commands: cli.Commands{
			{
				Name:  "summary",
				Usage: "Produce a JSON summary of all configured database schemas",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  flagPretty,
						Usage: "If provided, produces formatted JSON output",
					},
				},
				Action: summaryRun,
			},
			{
				Name:   "diff",
				Usage:  "Produce a diff of the database summaries",
				Action: diffRun,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(fmt.Sprintf("error during run: %v", err))
	}
}
