package main

import (
	"io"
	"slices"

	"github.com/urfave/cli/v2"
)

var (
	formatters map[string]FormatConstructor
)

func init() {
	formatters = map[string]FormatConstructor{
		FormatSimpleTable: newSimpleTableFormatter,
	}
}

func AvailableFormatters() []string {
	out := make([]string, len(formatters))
	i := 0
	for k := range formatters {
		out[i] = k
		i++
	}
	slices.Sort(out)
	return out
}

func BuildFormatter(cctx *cli.Context) (Formatter, error) {
	return formatters[cctx.String(flagFormat)](cctx, cctx.Value(flagFormatConfig).(MapString))
}

type Formatter interface {
	Type() string
	Render(connectionSummaries, io.Writer) error
}

type FormatConstructor func(*cli.Context, map[string]string) (Formatter, error)
