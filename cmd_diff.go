package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

func diffRun(cctx *cli.Context) error {
	conns, err := preRun(cctx)
	if err != nil {
		return err
	}

	defer conns.Close()

	summaries, err := summarizeConnections(cctx.Context, conns)
	if err != nil {
		return fmt.Errorf("error building summaries: %w", err)
	}

	tw := table.NewWriter()

	hdr := table.Row{""}
	for _, n := range summaries.DatabaseNames() {
		hdr = append(hdr, n)
	}

	tw.AppendHeader(hdr)

	for _, tn := range summaries.AllTableNames() {
		row := table.Row{tn}
		for _, cs := range summaries {
			for _, db := range cs.Databases {
				if _, ok := db.FindTable(tn); ok {
					row = append(row, "O")
				} else {
					row = append(row, "X")
				}
			}
		}
		tw.AppendRow(row)
	}

	fmt.Println(tw.Render())

	return nil
}
