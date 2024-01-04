package main

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

func summaryRun(cctx *cli.Context) error {
	conns, err := preRun(cctx)
	if err != nil {
		return err
	}

	defer conns.Close()

	summaries, err := summarizeConnections(cctx.Context, conns)
	if err != nil {
		return fmt.Errorf("error building summaries: %w", err)
	}

	var b []byte
	if cctx.Bool(flagPretty) {
		b, err = json.MarshalIndent(summaries, "", "  ")
	} else {
		b, err = json.Marshal(summaries)
	}

	if err != nil {
		return fmt.Errorf("error json-marshalling summary: %w", err)
	}

	fmt.Println(string(b))

	return nil
}
