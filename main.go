package main

import (
	"fmt"
	"os"
	"slices"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli/v2"
)

const (
	flagConn         = "conn"
	flagPretty       = "pretty"
	flagFormat       = "format"
	flagFormatConfig = "format-config"
	flagOut          = "out"
	flagOutConfig    = "out-config"
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
				Usage:    "A single MySQL connection with structure: \"addr=$addr user=$user pass=$pass db=$db[ db=$dbX][ label=$label]\"",
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
				Flags: []cli.Flag{
					// formatter and config
					&cli.StringFlag{
						Name:        flagFormat,
						Usage:       fmt.Sprintf("Formatter to use.  Available formatters: %v", AvailableFormatters()),
						Value:       FormatSimpleTable,
						DefaultText: FormatSimpleTable,
						Action: func(_ *cli.Context, v string) error {
							available := AvailableFormatters()
							if !slices.Contains(available, v) {
								return fmt.Errorf("unknown formatter %q specified, expected one of: %v", v, available)
							}
							return nil
						},
					},
					&MapStringFlag{
						Name:     flagFormatConfig,
						Usage:    `Configuration map for the specified formatter.  Available keys depend on formatter.  Must follow structure: "key=value,key2=value2"`,
						Required: false,
					},

					// output and config
					&cli.StringFlag{
						Name:        flagOut,
						Usage:       fmt.Sprintf("Destination of formatted diff.  Available outputs: %v", AvailableOutputs()),
						Value:       OutputStdOut,
						DefaultText: OutputStdOut,
						Action: func(_ *cli.Context, v string) error {
							available := AvailableOutputs()
							if !slices.Contains(available, v) {
								return fmt.Errorf("unknown output %q specified, expected to be one of: %v", v, available)
							}
							return nil
						},
					},
					&MapStringFlag{
						Name:     flagOutConfig,
						Usage:    `Configuration map for the specified output.  Available keys depend on output.  Must follow structure: "key=value,key2=value2"`,
						Required: false,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(fmt.Sprintf("error during run: %v", err))
	}
}
