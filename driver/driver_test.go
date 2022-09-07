package driver

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	// 例子：https://www.feishu.cn/base/bascnQIrLs6MrhIvftGsdYJgRFd
	appID       = os.Getenv("APP_ID")
	appSecret   = os.Getenv("APP_SECRET")
	appToken    = os.Getenv("APP_TOKEN")
	testTable1  = os.Getenv("TABLE_1")
	testTable2  = os.Getenv("TABLE_2")
	testUser1   = os.Getenv("USER_1")
	testRecord1 = os.Getenv("RECORD_1")
	testDSN     = fmt.Sprintf("bitable://%s:%s@open.feishu.cn/%s?log_level=trace", appID, appSecret, appToken)
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	logrus.Debug("register bitable driver")

	code := m.Run()

	// clean table2
	cleanTable(testTable2)
	os.Exit(code)
}

func cleanTable(table string) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		logrus.Fatalf("some error %s", err.Error())
	}

	rows, err := db.Query(fmt.Sprintf("select * from %s", table))
	res, err := loadRecord(rows)
	if err != nil {
		logrus.Fatalf("db query error: %s", err.Error())
	}
	for recordID := range res {
		_, err = db.Query(fmt.Sprintf("DELETE FROM %s WHERE record_id='%s'", testTable2, recordID))
	}
}

func TestShow(t *testing.T) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		t.Errorf("some error %s", err.Error())
	}

	t.Run("use database", func(t *testing.T) {
		_, err := db.Query(fmt.Sprintf("use %s", appToken))
		assert.NoError(t, err)
	})
	t.Run("show tables", func(t *testing.T) {
		rows, err := db.Query("show tables")
		assert.NoError(t, err)
		for rows.Next() {
			var tid, name string
			var version int
			if err := rows.Scan(&tid, &name, &version); err != nil {
				assert.NoError(t, err, "scan value error")
			} else {
				log.Println(tid, name, version)
			}
			assert.NotEmpty(t, tid)
		}
	})
	t.Run("show view", func(t *testing.T) {
		rows, err := db.Query(fmt.Sprintf("SHOW CREATE VIEW `%s`", testTable1))
		assert.NoError(t, err)
		for rows.Next() {
			var view, name, vType string
			if err := rows.Scan(&view, &name, &vType); err != nil {
				assert.NoError(t, err, "scan value error")
			} else {
				log.Println(view, name, vType)
			}
			assert.NotEmpty(t, view)
		}
	})
	t.Run("show columns", func(t *testing.T) {
		rows, err := db.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", testTable1))
		assert.NoError(t, err)
		for rows.Next() {
			var field, name, extra string
			var fType int
			if err := rows.Scan(&field, &fType, &name, &extra); err != nil {
				assert.NoError(t, err, "scan value error")
			} else {
				log.Println(field, name, fType, extra)
			}
		}
	})

}

func TestRecord(t *testing.T) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		t.Errorf("some error %s", err.Error())
	}

	equalHandleFunc := func(n int) func(t assert.TestingT, object interface{}, msgAndArgs ...interface{}) bool {
		return func(t assert.TestingT, object interface{}, msgAndArgs ...interface{}) bool {
			return assert.Equal(t, object, n, msgAndArgs...)
		}
	}
	tests := []struct {
		name          string
		sql           string
		wantAssertion assert.ValueAssertionFunc
		queryErr      assert.ErrorAssertionFunc
		loadErr       assert.ErrorAssertionFunc
	}{
		{
			name:          "select in",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE `数字` in (3, 1) order by `数字` desc limit 2", testTable1),
			wantAssertion: equalHandleFunc(2),
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select gt",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE `数字` >= 2 limit 10", testTable1),
			wantAssertion: assert.NotEmpty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select is not null",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE 单选 IS NOT NULL", testTable1),
			wantAssertion: assert.NotEmpty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select is null",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE 单选 IS NULL", testTable1),
			wantAssertion: assert.NotEmpty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select record_id not found",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE record_id='not found'", testTable1),
			wantAssertion: assert.Empty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select record_id",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE record_id='%s'", testTable1, testRecord1),
			wantAssertion: assert.NotEmpty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select date",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE `日期` >= TODATE('2021-12-16')", testTable1),
			wantAssertion: assert.NotEmpty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "select in",
			sql:           fmt.Sprintf("SELECT * FROM %s WHERE %s.`数字` = 1 order by %s.`数字` desc limit 1", testTable1, testTable1, testTable1),
			wantAssertion: assert.NotEmpty,
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
		{
			name:          "",
			sql:           fmt.Sprintf("SELECT * FROM %s limit 3", testTable1),
			wantAssertion: equalHandleFunc(3),
			queryErr:      assert.NoError,
			loadErr:       assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			tt.queryErr(t, err)
			res, err := loadRecord(rows)
			tt.loadErr(t, err)
			tt.wantAssertion(t, len(res))
		})
	}

}

func TestAlter(t *testing.T) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		t.Errorf("some error %s", err.Error())
	}
	t.Run("alter add/delete", func(t *testing.T) {
		_, err = db.QueryContext(context.Background(), fmt.Sprintf("ALTER TABLE %s DROP COLUMN `人员`;", testTable2))
		assert.NoError(t, err)
		// 这里使用 varchar/varbinary 里面的 数字表示字段类型
		// 多行文本 text|varchar(1)、数字 int|varchar(2)
		rows, err := db.Query(fmt.Sprintf("ALTER TABLE %s ADD COLUMN `人员` varchar(11);", testTable2))
		assert.NoError(t, err)
		assert.NotEmpty(t, rows.Next())
	})

	t.Run("alter change", func(t *testing.T) {
		_, err = db.Query(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN `日期` TO `测试修改`", testTable2))
		assert.NoError(t, err)
		_, err = db.Query(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN `测试修改` text;", testTable2))
		assert.NoError(t, err)
		_, err = db.Query(fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN `测试修改` `日期` varchar(5);", testTable2))
		assert.NoError(t, err)

	})
}

func TestTable(t *testing.T) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		t.Errorf("some error %s", err.Error())
	}

	t.Run("create", func(t *testing.T) {
		_, err := db.Query("CREATE TABLE `Lead` (`field_name` longtext,`field_type` longtext,`field_template_value_type` longtext,`template_name` longtext,`label` longtext,`length_limit` bigint,`address` boolean,`calc_formula` longtext,`calc_formula_type` longtext,`count_entity` longtext,`count_field` longtext,`count_relation_field` longtext,`currency` boolean,`currency_unit_type` longtext,`custom` boolean,`decimal_places` bigint,`default_to_zero` boolean,`default_value` longtext,`dict_name` longtext,`type_double` boolean,`encryption_required` boolean,`float_length_limit` bigint,`field_format` longtext,`type_formattable` boolean,`type_formula` boolean,`type_long_text` boolean,`mask` boolean,`max_length` bigint,`max_quantity` bigint,`number` boolean,`record_type` boolean,`required` boolean,`search_source_field_name` longtext,`validate_service_name_list_str` longtext,`value_type` longtext);")
		assert.NoError(t, err)
	})
	t.Run("create/delete table", func(t *testing.T) {
		tableName := "table_create"
		_, err := db.Query(fmt.Sprintf(`CREATE TABLE %s (PSM text, 选项 varchar(3) COMMENT '{"options":[{"name":"是"}]}', 多人协作 varchar(11) COMMENT '{"multiple":true}') COMMENT 'default_view_name';`, tableName))
		assert.NoError(t, err)

		tableID := getTable(t, db, tableName)[0]
		_, err = db.Query(fmt.Sprintf("DROP TABLE %s;", tableID))
		assert.NoError(t, err)
	})
}

func TestView(t *testing.T) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		t.Errorf("some error %s", err.Error())
	}

	t.Run("create/delete view", func(t *testing.T) {
		tableName := "table_view"
		_, err := db.Query(fmt.Sprintf("create table %s (ID text);", tableName))
		assert.NoError(t, err)

		tableID := getTable(t, db, tableName)[0]
		viewName := "view3"
		_, err = db.Query(fmt.Sprintf("CREATE VIEW kanban.%s AS SELECT * FROM %s", viewName, tableID))
		assert.NoError(t, err)
		viewID := ""
		rows, err := db.Query(fmt.Sprintf("SHOW CREATE VIEW `%s`", tableID))
		assert.NoError(t, err)
		found := 0
		for rows.Next() {
			var view, name, vType string
			if err := rows.Scan(&view, &name, &vType); err != nil {
				assert.NoError(t, err, "scan value error")
			} else {
				if name == viewName {
					viewID = view
					found++
				}
				log.Println(view, name, vType)
			}
		}
		assert.NotEmpty(t, found)

		_, err = db.Query(fmt.Sprintf("DROP VIEW %s.%s;", tableID, viewID))
		assert.NoError(t, err)

		_, err = db.Query(fmt.Sprintf("DROP TABLE %s;", tableID))
		assert.NoError(t, err)
	})
}

func TestModify(t *testing.T) {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		t.Errorf("some error %s", err.Error())
	}

	t.Run("insert records", func(t *testing.T) {
		rows, err := db.Query(fmt.Sprintf("INSERT INTO %s (`多行文本`) VALUES ('%s'), ('%s')", testTable2, genFCode(), genFCode()))
		count := 0
		for rows.Next() {
			count++
		}
		assert.NoError(t, err)
		assert.Equal(t, count, 2)
	})

	t.Run("insert numbers", func(t *testing.T) {
		rows, err := db.Query(fmt.Sprintf("INSERT INTO %s (`数字`) VALUES (3), (3.0), (3.3), (0.3), (.3)", testTable2))
		assert.NoError(t, err)
		assert.True(t, rows.Next())
	})

	t.Run("insert link/options/persons", func(t *testing.T) {
		url := `{"link":"https://www.google.com","text":"Google"}`
		options := `["选项一", "选项二"]`
		persons := fmt.Sprintf(`[{"id":"%s"}]`, testUser1)
		rows, err := db.Query(fmt.Sprintf("INSERT INTO %s (url.`超链接`, options.`多选`, persons.`人员`) VALUES ('%s', '%s', '%s')", testTable2, url, options, persons))
		assert.NoError(t, err)
		count := 0
		for rows.Next() {
			count++
		}
		assert.Equal(t, count, 1)
	})

	t.Run("update by record", func(t *testing.T) {
		fieldID, _, err := genRecord(t, db)
		assert.NoError(t, err)
		_, err = db.Query(fmt.Sprintf("UPDATE %s set `单选`='是' WHERE record_id='%s'", testTable2, fieldID))
		assert.NoError(t, err)
	})

	t.Run("delete record", func(t *testing.T) {
		fieldID, _, err := genRecord(t, db)
		_, err = db.Query(fmt.Sprintf("DELETE FROM %s WHERE record_id='%s'", testTable2, fieldID))
		assert.NoError(t, err)
	})

	t.Run("modify and delete", func(t *testing.T) {
		fCode := genFCode()
		// insert records
		_, err := db.Query(fmt.Sprintf("INSERT INTO %s (`多行文本`) VALUES ('%s')", testTable2, fCode))
		assert.NoError(t, err)
		// update by search
		_, err = db.Query(fmt.Sprintf("UPDATE %s set `单选`='是' WHERE `多行文本`='%s'", testTable2, fCode))
		assert.NoError(t, err)
		_, err = db.Query(fmt.Sprintf("UPDATE %s set `单选`='否' WHERE `多行文本`='%s'", testTable2, fCode))
		assert.NoError(t, err)
		// delete by search
		_, err = db.Query(fmt.Sprintf("DELETE FROM %s WHERE `多行文本`='%s' limit 1", testTable2, fCode))
		assert.NoError(t, err)
	})

}

func genRecord(t *testing.T, db *sql.DB) (string, string, error) {
	fCode := genFCode()
	_, err := db.Query(fmt.Sprintf("INSERT INTO %s (`多行文本`) VALUES ('%s')", testTable2, fCode))
	if err != nil {
		return "", "", err
	}
	rows, err := db.Query(fmt.Sprintf("SELECT `多行文本` FROM %s WHERE `多行文本`='%s' limit 1", testTable2, fCode))
	if err != nil {
		return "", "", err
	}
	res, err := loadRecord(rows)
	if err != nil {
		return "", "", err
	}
	for id, v := range res {
		return id, oneLine(v), nil
	}
	return "", "", err
}

func genFCode() string {
	return fmt.Sprintf("F%d", time.Now().UnixNano())
}

func getTable(t *testing.T, db *sql.DB, tableName string) []string {
	rows, err := db.Query("show tables;")
	assert.NoError(t, err)
	tables := []string{}
	for rows.Next() {
		var tid, name string
		var version int
		if err := rows.Scan(&tid, &name, &version); err != nil {
			assert.NoError(t, err, "scan value error")
		} else {
			if name == tableName {
				tables = append(tables, tid)
			}
		}
	}
	assert.NotEmpty(t, tables)
	return tables
}

func loadRecord(rows *sql.Rows) (map[string]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	records := make(map[string]map[string]interface{}, 0)
	for rows.Next() {
		fields := make([]interface{}, len(columns))
		scans := make([]interface{}, len(columns))
		for i := range fields {
			scans[i] = &fields[i]
		}
		err := rows.Scan(scans...)
		if err != nil {
			return nil, err
		}
		record := make(map[string]interface{})
		recordID := fields[0].(string)
		for i, col := range columns {
			record[col] = fields[i]
			fmt.Printf("%s: %v, ", col, fields[i])
		}
		fmt.Println("")
		records[recordID] = record
	}
	return records, nil
}
