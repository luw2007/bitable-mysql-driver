package driver

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

type Rows interface {
	Columns() []string
	Close() error
	Next(in Rows, dest []driver.Value) error
	Load() (pageList *lark.PageList, err error)
	Pick(dst []driver.Value, item interface{})
}

type rows struct {
	ctx      context.Context
	conn     *Conn
	appToken string
	columns  []string
	seek     int

	pageList *lark.PageList

	limit int64
	count int64
}

func newRows(ctx context.Context, conn *Conn, appToken string, columns []string, items []interface{}) *rows {
	return &rows{ctx: ctx, conn: conn, appToken: appToken, columns: columns, pageList: &lark.PageList{Items: items}}
}

func (r *rows) Clone(columns []string, items []interface{}) *rows {
	return &rows{
		ctx:      r.ctx,
		conn:     r.conn,
		appToken: r.appToken,
		columns:  columns,
		pageList: &lark.PageList{
			Items: items,
		},
	}
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Close() error {
	return nil
}

func (r *rows) Load() (*lark.PageList, error) {
	return &lark.PageList{}, nil
}

func (r *rows) Pick(dest []driver.Value, v interface{}) {
	if item, ok := v.([]interface{}); ok {
		for i, v := range item {
			dest[i] = v
		}
	}
}

func (r *rows) Next(in Rows, dest []driver.Value) error {
	for i := 0; i < maxLoopTimes; i++ {
		if r.seek < len(r.pageList.Items) {
			if r.limit > 0 && r.count >= r.limit {
				return io.EOF
			}
			in.Pick(dest, r.pageList.Items[r.seek])
			r.seek++
			r.count++
			return nil
		}
		// pageToken 判断是否加载过
		if r.pageList.PageToken != "" && !r.pageList.HasMore {
			return io.EOF
		}
		if err := r.loadMore(in); err != nil {
			if err == io.EOF {
				return err
			}
			return fmt.Errorf("loadMore data error: %w", err)
		}

	}
	return errors.New("too many times loading data")
}

func (r *rows) loadMore(in Rows) error {
	res, err := in.Load()
	if err != nil {
		return fmt.Errorf("load error: %w", err)
	}
	if res == nil || len(res.Items) == 0 {
		return io.EOF
	}
	r.seek = 0
	// 如果需要后续查询，pageToken 不应该为空
	r.pageList = res
	return nil
}

type rowsFactory struct {
	rows Rows
}

func newRowsFactory(rows Rows) *rowsFactory {
	return &rowsFactory{rows: rows}
}

func (l rowsFactory) Columns() []string {
	return l.rows.Columns()
}

func (l rowsFactory) Close() error {
	return l.rows.Close()
}

func (l rowsFactory) Next(dest []driver.Value) error {
	return l.rows.Next(l.rows, dest)
}
