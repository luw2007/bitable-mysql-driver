package driver

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/opcode"
	"github.com/pingcap/parser/test_driver"
	"github.com/pingcap/parser/types"
	"github.com/sirupsen/logrus"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

var (
	ErrNullValue = errors.New("null value")
)

// bitableStatement for sql statement
type bitableStatement struct {
	conn  *Conn
	ctx   context.Context
	stmt  []ast.StmtNode
	args  map[int]driver.NamedValue
	seek  int
	query string
}

// Close  implement for stmt
func (stmt *bitableStatement) Close() error {
	return nil
}

// QueryContext executes a query that may return rows
func (stmt *bitableStatement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	stmtNodes, _, err := stmt.conn.parser.Parse(stmt.query, "", "")
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] parser %w", err)
	}
	stmt.stmt = stmtNodes
	stmt.ctx = ctx
	stmt.args = buildNamedArgs(stmt.query, args)
	logrus.Debug("[bitable driver]  do query")
	baseRows := &rows{
		ctx:      stmt.ctx,
		conn:     stmt.conn,
		appToken: stmt.conn.AppToken,
	}
	if len(stmt.stmt) != 1 {
		return nil, fmt.Errorf("only one statement")
	}
	switch s := stmt.stmt[0].(type) {
	case *ast.UseStmt:
		return stmt.UseStmt(baseRows, s)
	case *ast.ShowStmt:
		return stmt.showStmt(baseRows, s)
	case *ast.SelectStmt:
		return stmt.selectStmt(baseRows, s)
	case *ast.CreateViewStmt:
		return stmt.createViewStmt(baseRows, s)
	case *ast.CreateTableStmt:
		return stmt.createTableStmt(baseRows, s)
	case *ast.DropTableStmt:
		return stmt.dropTableStmt(baseRows, s)
	case *ast.InsertStmt:
		return stmt.insertStmt(baseRows, s)
	case *ast.UpdateStmt:
		return stmt.updateStmt(baseRows, s)
	case *ast.DeleteStmt:
		return stmt.deleteStmt(baseRows, s)
	case *ast.AlterTableStmt:
		return stmt.alterTableStmt(baseRows, s)
	default:
		return nil, fmt.Errorf("bitable driver is not supported SQL: %s", stmt.query)
	}
}

func buildNamedArgs(query string, args []driver.NamedValue) map[int]driver.NamedValue {
	mask := '?'
	want := make(map[int]driver.NamedValue, len(args))
	index := 0
	for i, v := range query {
		if v == mask {
			want[i] = args[index]
			index++
		}
	}
	return want
}

func convertNamedValue(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, 0, len(args))
	for _, arg := range args {
		values = append(values, arg.Value)
	}
	return values
}

// Query  implement for Query
func (stmt *bitableStatement) Query(args []driver.Value) (driver.Rows, error) {
	panic("not implemented, use QueryContext instead")
}

func (stmt *bitableStatement) dropTableStmt(r *rows, s *ast.DropTableStmt) (driver.Rows, error) {
	for _, t := range s.Tables {
		table, view, err := stmt.getTableView(r.ctx, t)
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] get table view error: %w", err)
		}
		if s.IsView {
			if err := stmt.conn.DropView(r.ctx, r.appToken, table, view); err != nil {
				return nil, fmt.Errorf("[bitable driver] %w", err)
			}
			return nil, nil
		}
		if err := stmt.conn.DropTable(r.ctx, r.appToken, table); err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
	}
	return nil, nil
}

func (stmt *bitableStatement) alterTableStmt(r *rows, s *ast.AlterTableStmt) (driver.Rows, error) {
	table, _, err := stmt.getTableView(r.ctx, s.Table)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	for _, spec := range s.Specs {
		switch spec.Tp {
		case ast.AlterTableAddColumns:
			items := make([]interface{}, 0, len(spec.NewColumns))
			for _, column := range spec.NewColumns {
				fieldName := column.Name.Name.O
				fieldType := stmt.getFieldType(r.ctx, column.Tp)
				comment := stmt.getComment(column.Options)
				field, err := stmt.conn.AddField(r.ctx, r.appToken, table, fieldName, fieldType, comment)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] %w", err)
				}
				items = append(items, []interface{}{field.FieldID, field.FieldName, field.Type, oneLine(field.Property)})
			}
			columns := []string{"field_id", "name", "type", "property"}
			newRows := r.Clone(columns, items)
			return newRowsFactory(newRows), nil
		case ast.AlterTableDropColumn:
			fieldName := spec.OldColumnName.Name.O
			fieldId, _, _, err := stmt.getFieldID(r.ctx, r.appToken, table, fieldName)
			if err != nil {
				return nil, fmt.Errorf("[bitable driver] %w", err)
			}
			if _, err = stmt.conn.DeleteField(r.ctx, r.appToken, table, fieldId); err != nil {
				return nil, fmt.Errorf("[bitable driver] %w", err)
			}
			return nil, nil
		case ast.AlterTableRenameColumn:
			oldFieldName := spec.OldColumnName.Name.O
			fieldName := spec.NewColumnName.Name.O
			fieldID, oldFieldType, oldComment, err := stmt.getFieldID(r.ctx, r.appToken, table, oldFieldName)
			if err != nil {
				return nil, fmt.Errorf("[bitable driver] %w", err)
			}
			_, err = stmt.conn.UpdateField(r.ctx, r.appToken, table, fieldID, fieldName, oldFieldType, oldComment)
			if err != nil {
				return nil, fmt.Errorf("[bitable driver] %w", err)
			}
		case ast.AlterTableModifyColumn, ast.AlterTableChangeColumn:
			for _, column := range spec.NewColumns {
				fieldName := column.Name.Name.O
				fieldType := stmt.getFieldType(r.ctx, column.Tp)
				oldFieldName := fieldName
				// modify column
				if spec.OldColumnName != nil {
					oldFieldName = spec.OldColumnName.Name.O
				}
				fieldID, oldFieldType, oldComment, err := stmt.getFieldID(r.ctx, r.appToken, table, oldFieldName)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] %w", err)
				}
				if fieldType == 0 {
					fieldType = oldFieldType
				}
				comment := stmt.getComment(column.Options)
				if comment == "" {
					comment = oldComment
				}
				_, err = stmt.conn.UpdateField(r.ctx, r.appToken, table, fieldID, fieldName, fieldType, comment)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] %w", err)
				}
			}
		}
	}
	return nil, nil
}

func (stmt *bitableStatement) deleteStmt(r *rows, s *ast.DeleteStmt) (driver.Rows, error) {
	table, view, err := stmt.getTableView(r.ctx, s.TableRefs)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	filter, err := stmt.buildFilter(r.ctx, s.Where)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}

	var recordID string
	onlyRecordIDFilter := "CurrentValue.[record_id] = "
	if strings.HasPrefix(filter, onlyRecordIDFilter) {
		recordID = filter[len(onlyRecordIDFilter)+1 : len(filter)-1]
		filter = ""
	}

	deleteCallback := func(ctx context.Context, m map[string]map[string]interface{}) (int, error) {
		count := 0
		for recordID := range m {
			ok, err := stmt.conn.DeleteRecord(ctx, r.appToken, table, recordID)
			if err != nil {
				return 0, err
			}
			if ok {
				count++
			}
		}
		return count, nil
	}
	var limit int64
	if s.Limit != nil {
		if v, ok := s.Limit.Count.(*test_driver.ValueExpr); ok {
			limit = v.GetInt64()
		}
	}
	if len(recordID) > 0 {
		_, err = deleteCallback(r.ctx, map[string]map[string]interface{}{recordID: nil})
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
		return nil, nil
	}
	return stmt.searchRecords(r.ctx, r.appToken, table, view, filter, nil, limit, deleteCallback)
}

func (stmt *bitableStatement) updateStmt(r *rows, s *ast.UpdateStmt) (driver.Rows, error) {
	table, view, err := stmt.getTableView(r.ctx, s.TableRefs)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	filter, err := stmt.buildFilter(r.ctx, s.Where)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}

	var recordID string
	onlyRecordIDFilter := "CurrentValue.[record_id] = "
	if strings.HasPrefix(filter, onlyRecordIDFilter) {
		recordID = filter[len(onlyRecordIDFilter)+1 : len(filter)-1]
		filter = ""
	}
	updateCallBack := func(ctx context.Context, m map[string]map[string]interface{}) (int, error) {
		records, err := stmt.conn.UpdateRecords(ctx, r.appToken, table, m)
		if err != nil {
			return 0, err
		}
		return len(records), nil
	}

	var limit int64
	if s.Limit != nil {
		if v, ok := s.Limit.Count.(*test_driver.ValueExpr); ok {
			limit = v.GetInt64()
		}
	}

	data := make(map[string]interface{})
	for _, row := range s.List {
		switch v := row.Expr.(type) {
		case *test_driver.ValueExpr:
			data[row.Column.Name.O] = v.GetValue()
		case *test_driver.ParamMarkerExpr:
			if v2, ok := stmt.args[v.Offset]; ok {
				switch s := v2.Value.(type) {
				case string:
					data[row.Column.Name.O] = s
				default:
					data[row.Column.Name.O] = v2.Value
				}
			}
		}
	}
	if _, ok := data[FieldKeyRecordID]; ok {
		delete(data, FieldKeyRecordID)
	}
	if len(recordID) > 0 {
		_, err = updateCallBack(r.ctx, map[string]map[string]interface{}{recordID: data})
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
		return nil, nil
	}
	return stmt.searchRecords(r.ctx, r.appToken, table, view, filter, data, limit, updateCallBack)
}

func (stmt *bitableStatement) insertStmt(r *rows, s *ast.InsertStmt) (driver.Rows, error) {
	table, _, err := stmt.getTableView(r.ctx, s.Table)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}

	data := make([]map[string]interface{}, 0, len(s.Lists))
	for _, row := range s.Lists {
		record := make(map[string]interface{})
		for i, v := range row {
			fieldKey := s.Columns[i].Name.O
			// hack 'record_id'
			if fieldKey == FieldKeyRecordID {
				continue
			}
			var b []byte

			markerExpr, ok := v.(*test_driver.ParamMarkerExpr)
			if ok {
				switch vv := stmt.args[markerExpr.Offset].Value.(type) {
				case string:
					b = []byte(vv)
				case []byte:
					b = vv
				case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64:
					record[fieldKey] = vv
					continue
				case time.Time:
					record[fieldKey] = vv.UnixNano() / 1e6
					continue
				}
				if len(b) == 0 {
					record[fieldKey] = ""
					continue
				}
			} else {
				if tv, ok := v.(*test_driver.ValueExpr); ok {
					b = tv.Datum.GetBytes()
					if len(b) == 0 {
						// maybe int or decimal
						switch v.GetType().Tp {
						case mysql.TypeDecimal, mysql.TypeTiny, mysql.TypeShort, mysql.TypeLong, mysql.TypeLonglong,
							mysql.TypeTimestamp, mysql.TypeInt24:
							record[fieldKey] = v.(*test_driver.ValueExpr).Datum.GetInt64()
						case mysql.TypeFloat, mysql.TypeDouble:
							record[fieldKey] = v.(*test_driver.ValueExpr).Datum.GetFloat64()
						case mysql.TypeNewDecimal:
							d := v.(*test_driver.ValueExpr).Datum.GetMysqlDecimal().String()
							record[fieldKey], err = strconv.ParseFloat(d, 64)
							if err != nil {
								return nil, fmt.Errorf("parse record number field error: %w", err)
							}
						default:
							return nil, fmt.Errorf("not supported for  %v", v.GetType().Tp)
						}
						continue
					}
				}
			}

			recordKey := RecordKey(strings.ToLower(s.Columns[i].Table.O))
			switch recordKey {
			case RecordKeyUrl:
				var link RecordUrl
				err = json.Unmarshal(b, &link)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] parser url %w", err)
				}
				record[fieldKey] = link
			case RecordKeyAttachments:
				var attachments RecordAttachments
				err = json.Unmarshal(b, &attachments)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] parser attachments %w", err)
				}
				record[fieldKey] = attachments
			case RecordKeyOptions:
				var options RecordOptions
				err = json.Unmarshal(b, &options)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] parser options %w", err)
				}
				record[fieldKey] = options
			case RecordKeyPerson:
				var users RecordPersons
				err = json.Unmarshal(b, &users)
				if err != nil {
					return nil, fmt.Errorf("[bitable driver] parser persons %w", err)
				}
				record[fieldKey] = users
			default:
				if len(stmt.args) == 0 {
					record[fieldKey] = v.(*test_driver.ValueExpr).GetValue()
				} else {
					record[fieldKey] = string(b)
				}
			}

		}
		data = append(data, record)
	}
	if len(data) == 0 {
		return nil, errors.New("not found any record")
	}
	_, err = stmt.conn.InsertRecords(r.ctx, r.appToken, table, data)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	columns := make([]string, 0, len(s.Columns))
	for _, column := range s.Columns {
		columns = append(columns, column.Name.O)
	}
	items := make([]interface{}, 0, len(s.Lists))
	for _, record := range data {
		tmp := make([]interface{}, 0, len(record))
		for _, v := range record {
			tmp = append(tmp, v)
		}
		items = append(items, tmp)
	}
	newRows := r.Clone(columns, items)
	return newRowsFactory(newRows), nil
}

func (stmt *bitableStatement) createTableStmt(r *rows, s *ast.CreateTableStmt) (driver.Rows, error) {
	tableName, _, err := stmt.getTableView(r.ctx, s.Table)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	table, err := stmt.conn.CreateTable(r.ctx, r.appToken, tableName)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	// create view
	if viewName := stmt.getComment(s.Options); viewName != "" {
		viewNames, err := stmt.conn.ListViews(r.ctx, r.appToken, table, "", 10)
		oldViewID := viewNames.Items[0].(*lark.View).ViewID
		_, err = stmt.conn.CreateView(r.ctx, r.appToken, table, viewName, string(ViewTypeGrid))
		if err != nil {
			return nil, fmt.Errorf("create default view error: %w", err)
		}
		// 删除默认 view
		if err = stmt.conn.DropView(r.ctx, r.appToken, table, oldViewID); err != nil {
			return nil, fmt.Errorf("delete default view error: %w", err)
		}
	}
	// create columns
	for i, column := range s.Cols {
		fieldName := column.Name.Name.O
		fieldType := stmt.getFieldType(r.ctx, column.Tp)
		comment := stmt.getComment(column.Options)

		if i == 0 {
			fields, err := stmt.conn.ListFields(r.ctx, r.appToken, table, "", "", 1)
			if err != nil {
				return nil, fmt.Errorf("list default field error: %w", err)
			}
			oldFileID := fields.Items[0].(*lark.Field).FieldID
			if _, err := stmt.conn.UpdateField(r.ctx, r.appToken, table, oldFileID, fieldName, fieldType, comment); err != nil {
				return nil, fmt.Errorf("update field error: %w", err)
			}
			continue
		}
		if _, err := stmt.conn.AddField(r.ctx, r.appToken, table, fieldName, fieldType, comment); err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
	}
	columns := []string{"table"}
	items := []interface{}{[]interface{}{table}}
	newRows := r.Clone(columns, items)
	return newRowsFactory(newRows), nil
}

func (stmt *bitableStatement) createViewStmt(r *rows, s *ast.CreateViewStmt) (driver.Rows, error) {
	viewType, viewName, err := stmt.getTableView(r.ctx, s.ViewName)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	table, _, err := stmt.getTableView(r.ctx, s.Select.(*ast.SelectStmt).From)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	newView, err := stmt.conn.CreateView(r.ctx, r.appToken, table, viewName, strings.ToLower(viewType))
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}
	columns := []string{"id", "name", "type"}
	items := []interface{}{[]interface{}{newView.ViewID, newView.ViewName, newView.ViewType}}
	newRows := r.Clone(columns, items)
	return newRowsFactory(newRows), nil
}

func (stmt *bitableStatement) selectStmt(r *rows, s *ast.SelectStmt) (driver.Rows, error) {
	queryFields := stmt.buildFieldNames(r.ctx, s.Fields)
	if len(queryFields) == 1 && strings.ToLower(queryFields[0]) == "version()" {
		items := []interface{}{[]interface{}{biTableVersion}}
		newRows := r.Clone(queryFields, items)
		return newRowsFactory(newRows), nil
	}

	table, view, err := stmt.getTableView(r.ctx, s.From)
	if err != nil {
		return nil, fmt.Errorf("[bitable driver] %w", err)
	}

	var limit int64
	if s.Limit != nil {
		if v, ok := s.Limit.Count.(*test_driver.ValueExpr); ok {
			limit = v.GetInt64()
		}
	}

	filter, err := stmt.buildFilter(r.ctx, s.Where)
	if err != nil {
		return nil, fmt.Errorf("bitable driver filter error: %w", err)
	}
	// hack recordID
	var recordID string
	onlyRecordIDFilter := "CurrentValue.[record_id] = "
	if strings.HasPrefix(filter, onlyRecordIDFilter) {
		recordID = filter[len(onlyRecordIDFilter)+1 : len(filter)-1]
		filter = ""
	}
	sort := stmt.buildSort(r.ctx, s)

	fields, err := stmt.loadFields(r, table)
	if err != nil {
		return nil, err
	}
	if len(queryFields) == 0 {
		for k := range fields {
			queryFields = append(queryFields, k)
		}
	}

	return newRecordRows(r, table, view, sort, queryFields, fields, filter, recordID, limit), nil
}

func (stmt *bitableStatement) showStmt(r *rows, s *ast.ShowStmt) (driver.Rows, error) {
	switch s.Tp {
	case ast.ShowCreateDatabase:
		return newAppRows(r, r.appToken), nil
	case ast.ShowTables:
		return newTableRows(r), nil
	case ast.ShowCreateView:
		table, _, err := stmt.getTableView(r.ctx, s.Table)
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
		return newViewRows(r, table), nil
	case ast.ShowColumns:
		table, _, err := stmt.getTableView(r.ctx, s.Table)
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
		return newFieldRows(r, table, ""), nil
	}
	return nil, nil
}

func (stmt *bitableStatement) getComment(rawOption interface{}) (comment string) {
	switch options := rawOption.(type) {
	case []*ast.ColumnOption:
		for _, option := range options {
			if option.Tp == ast.ColumnOptionComment {
				comment = option.Expr.(*test_driver.ValueExpr).GetString()
			}
		}
	case []*ast.TableOption:
		for _, option := range options {
			if option.Tp == ast.TableOptionComment {
				comment = option.StrValue
			}
		}
	}

	return comment
}

// getFieldID 批量操作有性能问题
func (stmt *bitableStatement) getFieldID(ctx context.Context, appToken, table, fieldName string) (
	fieldId string, fieldType int64, property string, err error) {
	resp, err := stmt.conn.ListFields(ctx, appToken, table, "", "", 100)
	if err != nil {
		return "", 0, "", err
	}

	for _, item := range resp.Items {
		field := item.(*lark.Field)
		if field.FieldName == fieldName {
			fieldId = field.FieldID
			fieldType = field.Type
			if field.Property != nil {
				if b, err := json.Marshal(field.Property); err != nil {
					property = string(b)
				}
			}
			break
		}
	}
	if len(fieldId) == 0 {
		return "", 0, "", fmt.Errorf("not found fieldName: %s", fieldName)
	}
	return fieldId, fieldType, property, nil
}

func (stmt *bitableStatement) getTableView(ctx context.Context, node interface{}) (
	table string, view string, err error) {
	if node == nil {
		return
	}
	switch t := node.(type) {
	case *ast.TableRefsClause:
		if t != nil && t.TableRefs != nil && t.TableRefs.Left != nil {
			return stmt.getTableView(ctx, t.TableRefs.Left)
		}
	case *ast.TableSource:
		return stmt.getTableView(ctx, t.Source)
	case *ast.TableName:
		table = t.Schema.O
		view = t.Name.O
		if table == "" {
			table, view = view, ""
		}
		return table, view, nil
	}
	return "", "", errors.New("select not found table")
}

// NumInput row numbers
func (stmt *bitableStatement) NumInput() int {
	// don't know how many row numbers
	return -1
}

// Exec executes a query that doesn't return rows, such as an INSERT or UPDATE.
func (stmt *bitableStatement) Exec(args []driver.Value) (driver.Result, error) {
	panic("not implemented, use ExecContext")
}

// ExecContext executes a query that doesn't return rows, such as an INSERT or UPDATE.
func (stmt *bitableStatement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	rows, err := stmt.QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	res := bitableResult{}
	if rows == nil {
		return res, nil
	}
	for {
		v := make([]driver.Value, len(rows.Columns()))
		if err := rows.Next(v); err != nil {
			if io.EOF == err {
				break
			}
			res.rowsAffected++
			return nil, err
		}
	}
	return res, nil
	return stmt.Exec(convertNamedValue(args))
}

func (stmt *bitableStatement) searchRecords(ctx context.Context, appToken, table, view, filter string, fields map[string]interface{}, limit int64, callback func(context.Context, map[string]map[string]interface{}) (int, error)) (driver.Rows, error) {
	log := logrus.WithFields(logrus.Fields{
		"appToken":   appToken,
		"testTable2": table,
		"view":       view,
		"filter":     filter,
	})
	pageToken := ""
	count := int64(0)
	for i := 0; i < maxLoopTimes; i++ {
		pageSize := DefaultPageSize
		if limit > 0 && pageSize > limit {
			pageSize = limit
		}
		res, err := stmt.conn.ListRecords(ctx, appToken, table, view, "", filter, "", pageToken, pageSize)
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
		if res.Total == 0 {
			return nil, nil
		}
		pageToken = res.PageToken
		data := make(map[string]map[string]interface{}, len(res.Items))
		for _, item := range res.Items {
			record := item.(*lark.Record)
			count++
			if limit > 0 && limit <= count {
				continue
			}
			data[record.RecordID] = fields
		}
		if len(data) == 0 {
			break
		}
		n, err := callback(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("[bitable driver] %w", err)
		}
		count += int64(n)
		if !res.HasMore {
			return nil, nil
		}
	}
	log.Debugf("update %d records", count)
	if count == 0 {
		return nil, errors.New("not record changed")
	}
	return nil, nil
}

func (stmt *bitableStatement) getFieldType(_ context.Context, tp *types.FieldType) int64 {
	var fieldType FieldType
	switch tp.Tp {
	case mysql.TypeBlob, mysql.TypeTiny:
		fieldType = FieldTypeText
	case mysql.TypeDecimal, mysql.TypeShort, mysql.TypeLong, mysql.TypeLonglong,
		mysql.TypeFloat, mysql.TypeDouble, mysql.TypeInt24:
		fieldType = FieldTypeNumber
	case mysql.TypeLongBlob:
		fieldType = FieldTypeText
	case mysql.TypeVarchar, mysql.TypeVarString:
		if tp.Flen > 0 {
			fieldType = FieldType(tp.Flen)
		} else {
			fieldType = FieldTypeText
		}
	}
	return int64(fieldType)
}

func (stmt *bitableStatement) buildFieldNames(_ context.Context, node *ast.FieldList) []string {
	if node == nil {
		return nil
	}
	tmp := make([]string, 0, len(node.Fields))
	for _, f := range node.Fields {
		switch v := f.Expr.(type) {
		case *test_driver.ValueExpr:
			tmp = append(tmp, v.GetString())
		case *ast.ColumnNameExpr:
			tmp = append(tmp, v.Name.Name.O)
		default:
			tmp = append(tmp, f.Text())
		}
		if f.WildCard != nil {
			return nil
		}
	}
	return tmp
}

func (stmt *bitableStatement) buildSort(_ context.Context, s *ast.SelectStmt) string {
	if s.OrderBy == nil || len(s.OrderBy.Items) == 0 {
		return ""
	}
	tmp := make([]string, 0, len(s.OrderBy.Items))
	for _, i := range s.OrderBy.Items {
		if n, ok := i.Expr.(*ast.ColumnNameExpr); ok {
			name := n.Name.Name.O
			if name == `""` || len(name) == 0 {
				continue
			}
			if i.Desc {
				tmp = append(tmp, fmt.Sprintf("%s DESC", name))
			} else {
				tmp = append(tmp, fmt.Sprintf("%s ASC", name))
			}
		}
	}
	if len(tmp) == 0 {
		return ""
	}
	return oneLine(tmp)
}

func (stmt *bitableStatement) buildFilter(ctx context.Context, node interface{}) (string, error) {
	if node == nil {
		return "", nil
	}
	switch root := node.(type) {
	case *ast.BinaryOperationExpr:
		l, err := stmt.buildFilter(ctx, root.L)
		if err != nil {
			return "", err
		}
		r, err := stmt.buildFilter(ctx, root.R)
		if err != nil {
			// ignore null value
			if err != ErrNullValue {
				return "", err
			}
		}
		buff := bytes.NewBuffer(nil)
		switch root.Op {
		case opcode.LogicAnd, opcode.LogicOr:
			root.Op.Format(buff)
			buff.WriteByte('(')
			buff.WriteString(l)
			buff.WriteByte(',')
			buff.WriteString(r)
			buff.WriteByte(')')
			return buff.String(), nil
		case opcode.Not:
			return fmt.Sprintf("NOT(%s)", l), nil
		case opcode.GE, opcode.LE, opcode.EQ, opcode.NE, opcode.LT, opcode.GT,
			opcode.LeftShift, opcode.RightShift, opcode.Minus, opcode.And, opcode.Or, opcode.Mod, opcode.Xor, opcode.Div, opcode.Mul:
			buff.WriteString(l + " ")
			root.Op.Format(buff)
			buff.WriteString(" " + r)
			return buff.String(), nil
		case opcode.Plus:
			return fmt.Sprintf("%s%%2B%s", l, r), nil
		case opcode.IsNull:
			return fmt.Sprintf(`%s=""`, l), nil
		case opcode.In:
			return fmt.Sprintf(`%s.contains("%s")`, l, r), nil
		case opcode.BitNeg, opcode.IntDiv, opcode.LogicXor, opcode.NullEQ,
			opcode.Like, opcode.Case, opcode.Regexp, opcode.IsFalsity, opcode.IsTruth:
			return "", fmt.Errorf("not supported op %s", root.Op)
		}
	case *ast.PatternInExpr:
		in := make([]string, 0, len(root.List))
		for _, l := range root.List {
			v, err := stmt.buildFilter(ctx, l)
			if err != nil {
				return "", err
			}
			in = append(in, v)
		}
		v, err := stmt.buildFilter(ctx, root.Expr)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`%s.contains(%s)`, v, strings.Join(in, ",")), nil
	case *ast.ColumnNameExpr:
		return fmt.Sprintf(`CurrentValue.[%s]`, root.Name.Name.O), nil
	case *test_driver.ValueExpr:
		// 数字和字符串处理方式不相同
		switch root.Kind() {
		case test_driver.KindString:
			return fmt.Sprintf(`"%s"`, root.GetString()), nil
		case test_driver.KindBytes:
			return fmt.Sprintf(`"%s"`, string(root.GetBytes())), nil
		case test_driver.KindNull:
			return "", ErrNullValue
		default:
			return fmt.Sprintf(`%v`, root.GetValue()), nil
		}
	case *ast.IsNullExpr:
		v, err := stmt.buildFilter(ctx, root.Expr)
		if err != nil {
			return "", err
		}
		isNull := fmt.Sprintf(`%s=""`, v)
		if root.Not {
			isNull = fmt.Sprintf(`Not(%s)`, isNull)
		}
		return isNull, nil
	case *ast.FuncCallExpr:
		args := make([]string, 0, len(root.Args))
		for _, a := range root.Args {
			v, err := stmt.buildFilter(ctx, a)
			if err != nil {
				return "", err
			}
			args = append(args, v)
		}
		switch root.FnName.L {
		case "date":
			return fmt.Sprintf("DATE(%s, %s, %s)", args[0], args[1], args[2]), nil
		case "day", "month", "year", "todate":
			if len(args) != 1 {
				return "", errors.New("TODATE only accepts one argument")
			}
			return fmt.Sprintf("%s(%s)", root.FnName.O, args[0]), nil
		case "today":
			return fmt.Sprintf("TODAY()"), nil
		case "weekday":
			if len(root.Args) != 2 {
				return "", errors.New("TODATE only accepts two argument")
			}
			return fmt.Sprintf("WEEKDAY(%s, %s)", args[0], args[1]), nil
		}
	case *test_driver.ParamMarkerExpr:
		v, err := stmt.buildFilter(ctx, &root.ValueExpr)
		if err != nil {
			if err == ErrNullValue {
				v, ok := stmt.args[root.Offset]
				if ok {
					switch s := v.Value.(type) {
					case string:
						return fmt.Sprintf(`"%s"`, s), nil
					default:
						return fmt.Sprint(s), nil
					}
				}
			}
			return "", err
		}
		return v, nil
	default:
		logrus.Errorf("filter not supported %+v", node)
	}

	return "", nil
}

func (stmt *bitableStatement) loadFields(r *rows, table string) (map[string]lark.Field, error) {
	fields := make(map[string]lark.Field, 16)
	row := newFieldRows(r, table, "")
	cols := row.Columns()
	for {
		v := make([]driver.Value, len(cols))
		if err := row.Next(v); err != nil {
			if io.EOF == err {
				break
			}
			return nil, err
		}
		var p lark.FieldProperty
		if err := json.Unmarshal([]byte(v[3].(string)), &p); err != nil {
			return nil, fmt.Errorf("bitable %w", err)
		}
		fields[v[2].(string)] = lark.Field{
			FieldID:   v[0].(string),
			FieldName: v[2].(string),
			Type:      v[1].(int64),
			Property:  &p,
		}

	}
	return fields, nil
}

func (stmt *bitableStatement) UseStmt(r *rows, s *ast.UseStmt) (driver.Rows, error) {
	_, err := stmt.conn.GetApp(r.ctx, s.DBName)
	if err != nil {
		return nil, fmt.Errorf("change database[%s] error: %v", r.appToken, err)
	}
	r.conn.AppToken = s.DBName
	return nil, nil
}

func oneLine(obj interface{}) string {
	b, _ := json.Marshal(obj)
	return string(b)
}
