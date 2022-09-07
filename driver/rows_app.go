package driver

import (
	"database/sql/driver"
	"fmt"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

type appRows struct {
	*rows
	eof bool
}

func newAppRows(rows *rows, appToken string) driver.Rows {
	columns := []string{"app_token", "name", "revision"}
	newRows := rows.Clone(columns, nil)
	newRows.appToken = appToken
	return newRowsFactory(&appRows{rows: newRows})
}

func (a *appRows) Load() (*lark.PageList, error) {
	if a.eof {
		return &lark.PageList{}, nil
	}

	res, err := a.conn.GetApp(a.ctx, a.appToken)
	if err != nil {
		return nil, fmt.Errorf("change database[%s] error: %v", a.appToken, err)
	}
	a.eof = true
	return &lark.PageList{
		Items:     []interface{}{res},
		Total:     1,
		PageToken: loadOneTime,
	}, nil
}

func (a *appRows) Pick(dest []driver.Value, v interface{}) {
	item, ok := v.(*lark.AppMeta)
	if !ok {
		return
	}
	dest[0] = item.AppToken
	dest[1] = item.Name
	dest[2] = item.Revision
}
