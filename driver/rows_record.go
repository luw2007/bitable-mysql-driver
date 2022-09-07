package driver

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

const (
	FieldKeyRecordID = "record_id"
)

type recordRows struct {
	*rows
	table      string
	view       string
	sort       string
	fieldNames string
	filter     string
	fields     map[string]lark.Field
	recordID   string
}

func newRecordRows(base *rows, table string, view string, sort string, fieldNames []string,
	fields map[string]lark.Field, filter string, recordID string, limit int64) driver.Rows {
	newRows := base.Clone(nil, nil)
	newRows.columns = append([]string{FieldKeyRecordID}, fieldNames...)
	newRows.limit = limit
	newRows.pageList = &lark.PageList{}
	p := &recordRows{rows: newRows, table: table, view: view, sort: sort,
		fieldNames: oneLine(fieldNames), fields: fields, filter: filter, recordID: recordID}
	return newRowsFactory(p)
}

func (p *recordRows) Pick(dst []driver.Value, data interface{}) {
	item, ok := data.(*lark.Record)
	if !ok {
		return
	}
	dst[0] = item.RecordID
	for i, col := range p.columns {
		if v, ok := item.Fields[col]; ok {
			if f, ok := p.fields[col]; ok {
				switch FieldType(f.Type) {
				case FieldTypeText, FieldTypeSelect:
					dst[i] = v
				case FieldTypeNumber:
					if v == nil {
						dst[i] = 0
					} else {
						dst[i], _ = strconv.ParseFloat(v.(string), 64)
					}
				case FieldTypeCheckbox:
					dst[i], _ = v.(bool)
				case FieldTypeLink, FieldTypePerson, FieldTypeAttachment, FieldTypeMultipleSelect:
					dst[i] = oneLine(v)
				case FieldTypeDate, FieldTypeCreateTime, FieldTypeUpdateTime:
					if v != nil {
						dst[i] = time.Unix(int64(v.(float64)/1e3), 0)
					}
				}
				continue
			}
			dst[i] = oneLine(v)
		}
	}
}

func (p *recordRows) Load() (*lark.PageList, error) {
	pageSize := DefaultPageSize
	if p.limit > 0 && p.limit < pageSize {
		pageSize = p.limit
	}
	if p.recordID != "" {
		return p.loadRecord()
	}
	res, err := p.conn.ListRecords(p.ctx, p.appToken, p.table, p.view, p.fieldNames, p.filter, p.sort, p.pageList.PageToken, pageSize)
	if err != nil {
		return nil, fmt.Errorf("load records %w", err)
	}
	return res, nil
}

func (p recordRows) loadRecord() (*lark.PageList, error) {
	res, err := p.conn.GetRecord(p.ctx, p.appToken, p.table, p.recordID)
	if err != nil {
		return nil, fmt.Errorf("get record: %w", err)
	}
	items := []interface{}{res}
	return &lark.PageList{
		// 如果需要后续查询，pageToken 不应该为空
		PageToken: loadOneTime,
		Total:     1,
		HasMore:   false,
		Items:     items,
	}, nil
}
