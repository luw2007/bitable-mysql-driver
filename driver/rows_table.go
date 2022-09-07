package driver

import (
	"database/sql/driver"
	"fmt"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

type tableRows struct {
	*rows
}

func newTableRows(rows *rows) driver.Rows {
	newRows := rows.Clone(nil, nil)
	newRows.columns = []string{"id", "name", "revision"}
	return newRowsFactory(&tableRows{rows: newRows})
}

func (p *tableRows) Pick(dst []driver.Value, data interface{}) {
	item, ok := data.(*lark.Table)
	if !ok {
		return
	}
	dst[0] = item.TableID
	dst[1] = item.Name
	dst[2] = item.Revision
}

func (p *tableRows) Load() (*lark.PageList, error) {
	res, err := p.conn.ListTable(p.ctx, p.appToken, p.pageList.PageToken, DefaultPageSize)
	if err != nil {
		return nil, fmt.Errorf("load table error: %w", err)
	}

	return res, nil
}
