package driver

import (
	"context"
	"database/sql/driver"
	"errors"
)

type scannerConnection struct {
}

func (scannerConnection) Prepare(query string) (driver.Stmt, error) {
	return &scannerStatement{}, nil
}

func (scannerConnection) Close() error {
	return nil
}

func (scannerConnection) Begin() (driver.Tx, error) {
	return nil, nil
}

func (scannerConnection) Ping(ctx context.Context) error {
	return nil
}

func (scannerConnection) CheckNamedValue(*driver.NamedValue) error {
	return nil
}

type scannerStatement struct {
}

func (scannerStatement) CheckNamedValue(*driver.NamedValue) error {
	return nil
}

func (s scannerStatement) Close() error {
	return nil
}

func (s scannerStatement) NumInput() int {
	return 1
}

func (s scannerStatement) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("execution is not supported")
}

func (s scannerStatement) Query(args []driver.Value) (driver.Rows, error) {

	if len(args) < 1 {
		return nil, errors.New("scanner arguments should have an argument with rows")
	}

	rows, ok := args[0].(driver.Rows)
	if !ok {
		return nil, errors.New("scanner arguments should have an argument with rows")
	}

	return rows, nil
}
