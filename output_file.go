package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
)

const OutputFile = "file"

var _ Output = (*FileOutput)(nil)

type FileOutput struct {
	dest  string
	trunc bool
	mode  os.FileMode
}

func newFileOutput(_ *cli.Context, cfg map[string]string) (Output, error) {
	fo := FileOutput{
		mode:  os.FileMode(0666),
		trunc: true,
	}

	if dest, ok := cfg["dest"]; !ok {
		return nil, fmt.Errorf("output %q requires config key %q to be set", OutputFile, "dest")
	} else {
		fo.dest = dest
	}

	if trunc, ok := cfg["trunc"]; ok {
		if b, err := strconv.ParseBool(trunc); err != nil {
			return nil, fmt.Errorf("unable to parse \"trunc\" value %q as bool: %w", trunc, err)
		} else {
			fo.trunc = b
		}
	}

	if mode, ok := cfg["mode"]; ok {
		if i, err := strconv.ParseUint(mode, 8, 32); err != nil {
			return nil, fmt.Errorf("unable to parse \"mode\" value %q as int: %w", mode, err)
		} else {
			fo.mode = os.FileMode(i)
		}
	}

	return &fo, nil
}

func (*FileOutput) Type() string {
	return OutputFile
}

func (f *FileOutput) Writer() (io.Writer, error) {
	flgs := os.O_RDWR | os.O_CREATE

	if f.trunc {
		flgs |= os.O_TRUNC
	}

	w, err := os.OpenFile(f.dest, flgs, f.mode)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %w", f.dest, err)
	}
	return w, nil
}
