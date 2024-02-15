package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var _ cli.Flag = (*MapStringFlag)(nil)
var _ cli.VisibleFlag = (*MapStringFlag)(nil)
var _ cli.DocGenerationFlag = (*MapStringFlag)(nil)
var _ cli.CategorizableFlag = (*MapStringFlag)(nil)

// MapString is a flag implementation used to provide key=value semantics
// multiple times, or parse multiple keys out of a comma-separated string of "k1=v1,k2=v2[,]"
//
// Thanks, Frank.
type MapString map[string]string

func (ms *MapString) String() string {
	bits := make([]string, len(*ms))
	i := 0
	for k, v := range *ms {
		bits[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return strings.Join(bits, ",")
}

func (ms *MapString) Set(value string) error {
	if *ms == nil {
		*ms = make(MapString)
	}

	for _, s := range strings.Split(value, ",") {
		idx := strings.Index(value, "=")
		if idx == -1 {
			return fmt.Errorf("missing \"=\" value in key: %s", s)
		}
		(*ms)[s[0:idx]] = s[idx+1:]
	}

	return nil
}

func (ms *MapString) Get() any {
	return *ms
}

type MapStringFlag struct {
	Name string

	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value       *MapString
	Destination *MapString

	Aliases []string
	EnvVars []string

	defaultValue    map[string]string
	defaultValueSet bool

	Base map[string]string

	Action func(*cli.Context, map[string]string) error
}

func (f *MapStringFlag) String() string {
	return cli.FlagStringer(f)
}

func (f *MapStringFlag) Apply(set *flag.FlagSet) error {
	if f.Destination != nil && f.Value != nil {
		*f.Destination = make(map[string]string, len(*(f.Value)))
		for k, v := range *(f.Value) {
			(*(f.Destination))[k] = v
		}
	}

	var setValue *MapString
	switch {
	case f.Destination != nil:
		setValue = f.Destination
	case f.Value != nil:
		setValue = f.Value
	default:
		setValue = new(MapString)
	}

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if err := setValue.Set(val); err != nil {
			return fmt.Errorf("could not parse %q as map key=value pair from %s for flag %s: %w", val, source, f.Name, err)
		}

		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		set.Var(setValue, name, f.Usage)
	}

	return nil
}

func (f *MapStringFlag) Names() []string {
	return cli.FlagNames(f.Name, f.Aliases)
}

func (f *MapStringFlag) IsSet() bool {
	return f.HasBeenSet
}

func (f *MapStringFlag) GetCategory() string {
	return f.Category
}

func (f *MapStringFlag) TakesValue() bool {
	return true
}

func (f *MapStringFlag) GetUsage() string {
	return f.Usage
}

func (f *MapStringFlag) GetValue() string {
	if f.Value == nil {
		return ""
	}
	return f.Value.String()
}

func (f *MapStringFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

func (f *MapStringFlag) GetEnvVars() []string {
	return f.EnvVars
}

func (f *MapStringFlag) IsVisible() bool {
	return f.Hidden
}

// Return the first value from a list of environment variables and files
// (which may or may not exist), a description of where the value was found,
// and a boolean which is true if a value was found.
//
// Taken from https://github.com/urfave/cli/blob/7656c5fb838ca8a6febca43100147d317b544fd3/flag.go#L378
func flagFromEnvOrFile(envVars []string, filePath string) (value string, fromWhere string, found bool) {
	for _, envVar := range envVars {
		envVar = strings.TrimSpace(envVar)
		if value, found := os.LookupEnv(envVar); found {
			return value, fmt.Sprintf("environment variable %q", envVar), true
		}
	}
	for _, fileVar := range strings.Split(filePath, ",") {
		if fileVar != "" {
			if data, err := os.ReadFile(fileVar); err == nil {
				return string(data), fmt.Sprintf("file %q", filePath), true
			}
		}
	}
	return "", "", false
}
