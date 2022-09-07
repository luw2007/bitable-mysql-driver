package driver

import (
	"database/sql/driver"
	"fmt"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

type viewRows struct {
	*rows
	table string
}

func newViewRows(base *rows, table string) driver.Rows {
	newRows := base.Clone(nil, nil)
	newRows.columns = []string{"id", "name", "type"}
	return newRowsFactory(&viewRows{rows: newRows, table: table})
}

func (p *viewRows) Load() (*lark.PageList, error) {
	res, err := p.conn.ListViews(p.ctx, p.appToken, p.table, p.pageList.PageToken, DefaultPageSize)
	if err != nil {
		return nil, fmt.Errorf("load view rows: %w", err)
	}
	return res, nil
}

func (p *viewRows) Pick(dst []driver.Value, i interface{}) {
	item, ok := i.(*lark.View)
	if !ok {
		return
	}
	dst[0] = item.ViewID
	dst[1] = item.ViewName
	dst[2] = item.ViewType
}
