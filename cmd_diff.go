package main

import (
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
)

func diffRun(cctx *cli.Context) error {
	conns, err := preRun(cctx)
	if err != nil {
		return err
	}

	defer conns.Close()

	formatter, err := BuildFormatter(cctx)
	if err != nil {
		return fmt.Errorf("error building formatter: %w", err)
	}

	output, err := BuildOutput(cctx)
	if err != nil {
		return fmt.Errorf("error building output: %w", err)
	}

	outputWriter, err := output.Writer()
	if err != nil {
		return fmt.Errorf("error opening writer for output: %w", err)
	}

	// if this is a closeable writer, queue up close.
	if wc, ok := outputWriter.(io.Closer); ok {
		defer func() { _ = wc.Close() }()
	}

	summaries, err := summarizeConnections(cctx.Context, conns)
	if err != nil {
		return fmt.Errorf("error building summaries: %w", err)
	}

	return formatter.Render(summaries, outputWriter)
}
