package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	keyLabel = "label"
	keyAddr  = "addr"
	keyUser  = "user"
	keyPass  = "pass"
	keyDB    = "db"
)

type connConfig struct {
	Label     string
	Address   string
	Username  string
	Password  string
	Databases []string
}

func parseConnConfig(in string) (connConfig, error) {
	var cc connConfig

	cs := strings.Split(in, " ")
	for _, s := range cs {
		s = strings.TrimSpace(s)
		p := strings.Split(s, "=")
		if len(p) != 2 {
			return connConfig{}, errors.New("connection format must be $key=$value")
		}

		key, value := strings.TrimSpace(p[0]), strings.TrimSpace(p[1])

		if key == "" {
			return connConfig{}, errors.New("key must not be empty")
		}
		if value == "" {
			continue
		}

		switch key {
		case keyLabel:
			if cc.Label != "" {
				return connConfig{}, fmt.Errorf("each conn must have only one %q key", keyLabel)
			}
			cc.Label = value
		case keyAddr:
			if cc.Address != "" {
				return connConfig{}, fmt.Errorf("each conn must have only one %q key", keyAddr)
			}
			cc.Address = value
		case keyUser:
			if cc.Username != "" {
				return connConfig{}, fmt.Errorf("each conn must have only one %q key", keyUser)
			}
			cc.Username = value
		case keyPass:
			if cc.Password != "" {
				return connConfig{}, fmt.Errorf("each conn must have only one %q key", keyPass)
			}
			cc.Password = value
		case keyDB:
			cc.Databases = append(cc.Databases, value)

		default:
			return connConfig{}, fmt.Errorf("uknown key %q", key)
		}
	}

	if cc.Address == "" {
		return connConfig{}, errors.New("address must not be empty")
	}
	if cc.Username == "" {
		return connConfig{}, errors.New("username must not be empty")
	}
	if cc.Password == "" {
		return connConfig{}, errors.New("password must not be empty")
	}
	if len(cc.Databases) == 0 {
		return connConfig{}, errors.New("must provide at least one database per connection config")
	}

	return cc, nil
}

func parseConnFlags(cctx *cli.Context) ([]connConfig, error) {
	var out []connConfig

	// fetch flag
	connFlags := cctx.StringSlice(flagConn)
	if len(connFlags) == 0 {
		return nil, errors.New("at least one connection configuration must be provided")
	}

	for _, cf := range connFlags {
		cc, err := parseConnConfig(cf)
		if err != nil {
			return nil, fmt.Errorf("error parsing connection config: %w", err)
		}
		out = append(out, cc)
	}

	return out, nil
}
