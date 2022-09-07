package driver

import (
	"database/sql/driver"
	"fmt"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

type fieldRows struct {
	*rows
	table string
	view  string
}

func newFieldRows(base *rows, table string, view string) driver.Rows {
	newRows := base.Clone(nil, nil)
	newRows.columns = []string{"id", "type", "name", "extra"}
	return newRowsFactory(&fieldRows{rows: newRows, table: table, view: view})
}

func (p *fieldRows) Pick(dst []driver.Value, i interface{}) {
	item, ok := i.(*lark.Field)
	if !ok {
		return
	}
	dst[0] = item.FieldID
	dst[1] = item.Type
	dst[2] = item.FieldName
	dst[3] = oneLine(item.Property)
}

func (p *fieldRows) Load() (*lark.PageList, error) {
	res, err := p.conn.ListFields(p.ctx, p.appToken, p.table, p.view, p.pageList.PageToken, DefaultPageSize)
	if err != nil {
		return nil, fmt.Errorf("load field list: %w", err)
	}
	return res, nil
}
