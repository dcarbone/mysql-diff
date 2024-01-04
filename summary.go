package main

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
)

type columnSummary struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Nullable string         `json:"nullable"`
	Key      string         `json:"key"`
	Default  sql.NullString `json:"default"`
	Extra    string         `json:"extra"`
}

type tableSummary struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Columns []columnSummary `json:"columns"`
}

func (ts tableSummary) FindColumn(name string) (columnSummary, bool) {
	for _, c := range ts.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return columnSummary{}, false
}

type databaseSummary struct {
	Name   string          `json:"name"`
	Tables []*tableSummary `json:"tables"`
}

func (ds databaseSummary) TableNames() []string {
	out := make([]string, 0)
	for _, tbl := range ds.Tables {
		out = append(out, tbl.Name)
	}
	return out
}

func (ds databaseSummary) FindTable(name string) (tableSummary, bool) {
	for _, tbl := range ds.Tables {
		if tbl.Name == name {
			return *tbl, true
		}
	}
	return tableSummary{}, false
}

type connectionSummary struct {
	Label     string             `json:"label"`
	Address   string             `json:"address"`
	Databases []*databaseSummary `json:"databases"`
}

func (cs connectionSummary) DisplayName() string {
	if cs.Label != "" {
		return cs.Label
	}
	return cs.Address
}

type connectionSummaries []*connectionSummary

func (cs connectionSummaries) DatabaseNames() []string {
	out := make([]string, 0)
	for _, c := range cs {
		for _, db := range c.Databases {
			out = append(out, fmt.Sprintf("`%s`.`%s`", c.DisplayName(), db.Name))
		}
	}
	return out
}

func (cs connectionSummaries) AllTableNames() []string {
	out := make([]string, 0)
	for _, c := range cs {
		for _, db := range c.Databases {
			tblNames := db.TableNames()
			for _, tn := range tblNames {
				if !slices.Contains(out, tn) {
					out = append(out, tn)
				}
			}
		}
	}

	slices.Sort(out)

	return out
}

func addColumnSummaries(ctx context.Context, conn *sql.DB, db string, tblsum *tableSummary) error {
	tx, err := startTx(ctx, conn, db)
	if err != nil {
		return err
	}

	// always queue up rollback
	defer func() { _ = tx.Rollback() }()

	// use specific database
	_, err = doExec(ctx, tx, fmt.Sprintf("USE `%s`;", db))
	if err != nil {
		return err
	}

	rows, err := doQuery(ctx, tx, fmt.Sprintf("SHOW COLUMNS FROM `%s`;", tblsum.Name))
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var colsum columnSummary

		err = rows.Scan(&colsum.Name, &colsum.Type, &colsum.Nullable, &colsum.Key, &colsum.Default, &colsum.Extra)
		if err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		tblsum.Columns = append(tblsum.Columns, colsum)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func summarizeDatabase(ctx context.Context, conn *sql.DB, db string) (*databaseSummary, error) {
	tx, err := startTx(ctx, conn, db)
	if err != nil {
		return nil, err
	}

	// always queue up rollback
	defer func() { _ = tx.Rollback() }()

	// list all tables
	rows, err := doQuery(ctx, tx, "SHOW FULL TABLES;")
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	dbsum := &databaseSummary{
		Name:   db,
		Tables: make([]*tableSummary, 0),
	}

	for rows.Next() {
		var tname, ttype string

		if err = rows.Scan(&tname, &ttype); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		dbsum.Tables = append(dbsum.Tables, &tableSummary{
			Name:    tname,
			Type:    ttype,
			Columns: make([]columnSummary, 0),
		})
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error closing transaction: %w", err)
	}

	for _, tbl := range dbsum.Tables {
		if err = addColumnSummaries(ctx, conn, db, tbl); err != nil {
			return nil, fmt.Errorf("error summarizing database %q table %q columns: %w", db, tbl.Name, err)
		}
	}

	return dbsum, nil
}

func summarizeConnections(ctx context.Context, conns []*mysqlConn) (connectionSummaries, error) {
	summaries := make(connectionSummaries, 0)

	for _, cn := range conns {
		connStruct := &connectionSummary{
			Label:     cn.Label,
			Address:   cn.Address,
			Databases: make([]*databaseSummary, 0),
		}
		summaries = append(summaries, connStruct)
		for _, db := range cn.Databases {
			dbStruct, err := summarizeDatabase(ctx, cn.Conn, db)
			if err != nil {
				return nil, fmt.Errorf("error summarizing database %q in server %q: %w", db, cn.Address, err)
			}
			connStruct.Databases = append(connStruct.Databases, dbStruct)
		}
	}

	return summaries, nil
}
