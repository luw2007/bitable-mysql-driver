package driver

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/pingcap/parser"
	"github.com/stretchr/testify/assert"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

func TestConn_QueryContext(t *testing.T) {
	type fields struct {
		BiTable   *lark.BiTable
		parser    *parser.Parser
		AppID     string
		AppSecret string
		AppToken  string
	}
	type args struct {
		ctx   context.Context
		query string
		args  []driver.NamedValue
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Rows
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				BiTable:   tt.fields.BiTable,
				parser:    tt.fields.parser,
				AppID:     tt.fields.AppID,
				AppSecret: tt.fields.AppSecret,
				AppToken:  tt.fields.AppToken,
			}
			got, err := c.QueryContext(tt.args.ctx, tt.args.query, tt.args.args)
			if !tt.wantErr(t, err, fmt.Sprintf("QueryContext(%v, %v, %v)", tt.args.ctx, tt.args.query, tt.args.args)) {
				return
			}
			assert.Equalf(t, tt.want, got, "QueryContext(%v, %v, %v)", tt.args.ctx, tt.args.query, tt.args.args)
		})
	}
}
