package driver

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/pingcap/parser"
	"github.com/sirupsen/logrus"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

// WithUserToken add `user_access_token` to context.Value
func WithUserToken(ctx context.Context, userToken string) context.Context {
	return lark.WithUserToken(ctx, userToken)
}

// Conn for db open
type Conn struct {
	*lark.BiTable
	parser    *parser.Parser
	AppID     string
	AppSecret string
	AppToken  string
}

// Ping check client connection
func (c *Conn) Ping(ctx context.Context) error {
	_, err := c.GetApp(ctx, c.AppToken)
	return err
}

// QueryContext the context timeout and return when the context is canceled.
func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	stmt, err := c.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	return stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
}

func (c *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	statement, err := c.Prepare(query)
	if err != nil {
		return nil, nil
	}

	return statement.Query(args)
}

// Prepare statement for prepare exec
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

// PrepareContext statement for prepare exec
func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	logrus.Debugf("[bitable driver]  do prepare, sql: %s", query)

	return &bitableStatement{
		conn:  c,
		ctx:   ctx,
		query: query,
	}, nil
}

// Close db connection close
func (c *Conn) Close() error {
	return errors.New("can't close connection")
}

// Begin tx begin
func (c *Conn) Begin() (driver.Tx, error) {
	return &biTableTransaction{c}, nil
}
