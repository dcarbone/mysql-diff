package main

import (
	"fmt"
	"io"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

const (
	FormatSimpleTable = "simple-table"
)

var (
	_ Formatter = (*SimpleTableFormatter)(nil)

	styleMap = map[string]table.Style{
		"default": table.StyleDefault,
		"bold":    table.StyleBold,
		"double":  table.StyleDouble,
		"light":   table.StyleLight,
		"rounded": table.StyleRounded,

		"bright":                 table.StyleColoredBright,
		"dark":                   table.StyleColoredDark,
		"black-on-blue-white":    table.StyleColoredBlackOnBlueWhite,
		"black-on-cyan-white":    table.StyleColoredBlackOnCyanWhite,
		"black-on-green-white":   table.StyleColoredBlackOnGreenWhite,
		"black-on-magenta-white": table.StyleColoredBlackOnMagentaWhite,
		"black-on-yellow-white":  table.StyleColoredBlackOnYellowWhite,
		"black-on-red-white":     table.StyleColoredBlackOnRedWhite,
		"blue-white-on-black":    table.StyleColoredBlueWhiteOnBlack,
		"cyan-white-on-black":    table.StyleColoredCyanWhiteOnBlack,
		"green-white-on-black":   table.StyleColoredGreenWhiteOnBlack,
		"magenta-white-on-black": table.StyleColoredMagentaWhiteOnBlack,
		"red-white-on-black":     table.StyleColoredRedWhiteOnBlack,
		"yellow-white-on-black":  table.StyleColoredYellowWhiteOnBlack,
	}
)

func styleKeys() []string {
	out := make([]string, len(styleMap))
	i := 0
	for k := range styleMap {
		out[i] = k
		i++
	}
	return out
}

type SimpleTableFormatter struct {
	style table.Style

	header bool
}

func newSimpleTableFormatter(_ *cli.Context, cfg map[string]string) (Formatter, error) {
	var err error

	st := SimpleTableFormatter{
		style:  table.StyleDefault,
		header: true,
	}

	if v, ok := cfg["header"]; ok {
		if st.header, err = strconv.ParseBool(v); err != nil {
			return nil, fmt.Errorf("error parsing flag \"header\" value %q as bool: %w", v, err)
		}
	}

	if v, ok := cfg["style"]; ok {
		if st.style, ok = styleMap[v]; !ok {
			return nil, fmt.Errorf("unknown style %q specified, expected one of %v", v, styleKeys())
		}
	}

	return &st, nil
}

func (*SimpleTableFormatter) Type() string {
	return FormatSimpleTable
}

func (to *SimpleTableFormatter) Render(summaries connectionSummaries, sink io.Writer) error {
	tw := table.NewWriter()

	tw.SetStyle(to.style)

	hdr := table.Row{}
	for _, n := range summaries.DatabaseNames() {
		hdr = append(hdr, n)
	}

	if to.header {
		tw.AppendHeader(hdr)
	}

	for _, tn := range summaries.AllTableNames() {
		row := table.Row{}
		for _, cs := range summaries {
			for _, db := range cs.Databases {
				if _, ok := db.FindTable(tn); ok {
					row = append(row, tn)
				} else {
					row = append(row, "")
				}
			}
		}
		tw.AppendRow(row)
	}

	if _, err := sink.Write([]byte(tw.Render())); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}

	return nil
}
