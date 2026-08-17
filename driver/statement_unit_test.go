package driver

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/test_driver"
	"github.com/stretchr/testify/require"
)

func TestBuildFilterEscapesStringParameter(t *testing.T) {
	query := "SELECT * FROM t WHERE name = ?"
	nodes, _, err := parser.New().Parse(query, "", "")
	require.NoError(t, err)

	where := nodes[0].(*ast.SelectStmt).Where
	marker := where.(*ast.BinaryOperationExpr).R.(*test_driver.ParamMarkerExpr)
	stmt := bitableStatement{args: map[int]driver.NamedValue{
		marker.Offset: {Ordinal: 1, Value: `x") OR true OR ("`},
	}}

	filter, err := stmt.buildFilter(context.Background(), where)
	require.NoError(t, err)
	require.Equal(t, `CurrentValue.[name] = "x\") OR true OR (\""`, filter)
}

func TestQuoteFilterStringEscapesControlCharacters(t *testing.T) {
	require.Equal(t, `"line\npath\\name\t\"quoted\""`, quoteFilterString("line\npath\\name\t\"quoted\""))
}

func TestQuoteFilterStringPreservesPunctuation(t *testing.T) {
	require.Equal(t, `"AT&T a<b c>d"`, quoteFilterString("AT&T a<b c>d"))
}

func TestBindNamedArgsRejectsWrongCount(t *testing.T) {
	tests := []struct {
		name  string
		query string
		args  []driver.NamedValue
	}{
		{name: "missing", query: "SELECT * FROM t WHERE name = ?"},
		{name: "extra", query: "SELECT * FROM t", args: []driver.NamedValue{{Ordinal: 1, Value: "extra"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, _, err := parser.New().Parse(tt.query, "", "")
			require.NoError(t, err)
			_, err = bindNamedArgs(nodes, tt.args)
			require.ErrorContains(t, err, "expected")
		})
	}
}

func TestBindNamedArgsIgnoresQuestionMarkInStringLiteral(t *testing.T) {
	query := "SELECT * FROM t WHERE name = '?' AND value = ?"
	nodes, _, err := parser.New().Parse(query, "", "")
	require.NoError(t, err)
	marker := nodes[0].(*ast.SelectStmt).Where.(*ast.BinaryOperationExpr).R.(*ast.BinaryOperationExpr).R.(*test_driver.ParamMarkerExpr)

	args, err := bindNamedArgs(nodes, []driver.NamedValue{{Ordinal: 1, Value: "value"}})
	require.NoError(t, err)
	require.Equal(t, "value", args[marker.Offset].Value)
}

func TestQueryContextRejectsMissingParameter(t *testing.T) {
	conn := Conn{parser: parser.New()}
	_, err := conn.QueryContext(
		context.Background(),
		"SELECT * FROM t WHERE name = ?",
		nil,
	)
	require.ErrorContains(t, err, "expected 1 arguments, got 0")
}
