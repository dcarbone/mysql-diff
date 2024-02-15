package main

import (
	"io"
	"os"
	"slices"

	"github.com/urfave/cli/v2"
)

const (
	OutputStdOut = "stdout"
	OutputStdErr = "stderr"
)

var (
	stdOut = new(StdOutOutput)
	stdErr = new(StdErrOutput)

	outputs map[string]OutputConstructor
)

func init() {
	outputs = map[string]OutputConstructor{
		OutputStdOut: func(_ *cli.Context, _ map[string]string) (Output, error) { return stdOut, nil },
		OutputStdErr: func(_ *cli.Context, _ map[string]string) (Output, error) { return stdErr, nil },
		OutputFile:   newFileOutput,
	}
}

func AvailableOutputs() []string {
	out := make([]string, len(outputs))
	i := 0
	for k := range outputs {
		out[i] = k
		i++
	}
	slices.Sort(out)
	return out
}

func BuildOutput(cctx *cli.Context) (Output, error) {
	return outputs[cctx.String(flagOut)](cctx, cctx.Value(flagOutConfig).(MapString))
}

type Output interface {
	Type() string
	Writer() (io.Writer, error)
}

type OutputConstructor func(*cli.Context, map[string]string) (Output, error)

var (
	_ Output = (*StdOutOutput)(nil)
	_ Output = (*StdErrOutput)(nil)
)

type StdOutOutput struct{}

func (*StdOutOutput) Type() string               { return OutputStdOut }
func (*StdOutOutput) Writer() (io.Writer, error) { return os.Stdout, nil }

type StdErrOutput struct{}

func (*StdErrOutput) Type() string               { return OutputStdErr }
func (*StdErrOutput) Writer() (io.Writer, error) { return os.Stderr, nil }
