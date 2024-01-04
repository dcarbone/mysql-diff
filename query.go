package main

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/go-sql-driver/mysql"
)

type connOrTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type mysqlConn struct {
	Conn      *sql.DB
	Label     string
	Address   string
	Databases []string
}

type mysqlConns []*mysqlConn

func (mcs mysqlConns) Close() {
	for _, mc := range mcs {
		if mc.Conn != nil {
			_ = mc.Conn.Close()
		}
	}
}

func openConnections(connConfigs []connConfig) (mysqlConns, error) {
	conns := make(mysqlConns, 0)

	for _, cc := range connConfigs {
		// build mysql config from input
		mcfg := mysql.NewConfig()
		mcfg.Addr = cc.Address
		mcfg.User = cc.Username
		mcfg.Passwd = cc.Password

		// attempt to open connection
		mconn, err := mysql.NewConnector(mcfg)
		if err != nil {
			return nil, fmt.Errorf("error opening mysql connector: %w", err)
		}

		conn := &mysqlConn{
			Conn:      sql.OpenDB(mconn),
			Label:     cc.Label,
			Address:   cc.Address,
			Databases: slices.Clone(cc.Databases),
		}

		conns = append(conns, conn)
	}

	return conns, nil
}

func doExec(ctx context.Context, conn connOrTX, query string, params ...any) (sql.Result, error) {
	res, err := conn.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("error executing query %q: %w", query, err)
	}
	return res, nil
}

func doQuery(ctx context.Context, conn connOrTX, query string, params ...any) (*sql.Rows, error) {
	rows, err := conn.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("error fetching rows for query %q: %w", query, err)
	}
	return rows, nil
}

// startTx starts a read-only transaction, executing a USE statement with the provided database name.
func startTx(ctx context.Context, conn *sql.DB, db string) (*sql.Tx, error) {
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// use specific database
	_, err = doExec(ctx, tx, fmt.Sprintf("USE `%s`;", db))
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return tx, nil
}
