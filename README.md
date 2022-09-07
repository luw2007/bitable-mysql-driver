# bitable-mysql-driver

Feishu open platform provides bi-table api. Changed the sdk as a mysql driver for quick start.

[简体中文 README](./readme_cn.md)

## quick start

```
# dsn = "bitable://<app_id>:<app_secret>@open.feishu.cn/<app_token>?log_level=trace"
go run cmd/bsql/main.go 'bitable://cli_a14eda43cb7ad013:l5zyi***********************16Y0@open.feishu.cn/bascnQIrLs6MrhIvftGsdYJgRFd'

> show tables;
id: tblTqyMTqUTFrDQc
name: table1
revision: 3

id: tblebGSJc65Km9qG
name: table2
revision: 5

> show create view tblTqyMTqUTFrDQc;
id: vewbe3eCpw
name: Grid
type: grid

> show columns from tblebGSJc65Km9qG;
field: fldACpt0Hp
type: 1
comment: Text
extra: null

field: fld5Iuk6lB
type: 3
comment: Select
extra: {"options":[{"name":"option_one","id":"opt832qE9t"}]}

field: fld0ItJzco
type: 11
comment: Person
extra: {"multiple":true}
```

## Useful SQL

```
SHOW TABLES;
SHOW COLUMNS FROM table;
SHOW CREATE VIEW table;

# Select
SELECT * FROM table limit 10;
SELECT * FROM table WHERE `Number` >= 2 and `Person` in ('XX') limit 10;
SELECT * FROM table WHERE `Date` >= TODATE('2021-12-16');
SELECT * FROM table WHERE `Number` in (3, 1) order by `Number` desc limit 10;
SELECT * FROM table WHERE record_id = 'rec9eOiv5d';
SELECT * FROM table WHERE `Select` IS NOT NULL;
SELECT * FROM table WHERE `Select` IS NULL;


# DML
CREATE TABLE table
(
    `Text`   text,
    `Select` varchar(3) COMMENT '{"options":[{"name":"optione_one"}]}',
    `Person` varchar(11) COMMENT '{"multiple":true}'
) COMMENT 'Grid';
CREATE VIEW kanban.`kanban` AS SELECT * FROM table;
DROP TABLE table;

# DDL
ALTER TABLE table ADD COLUMN `Text` varchar(1) COMMENT '{"multiple":true}';
ALTER TABLE table CHANGE COLUMN `NewDate` `Date` varchar(5);
ALTER TABLE table MODIFY COLUMN `NewDate` text;
ALTER TABLE table RENAME COLUMN `Date` TO `NewDate`;
ALTER TABLE table DROP COLUMN `Date`;

# Records
INSERT INTO table (`Number`) VALUES (3), (3.0), (0.3), (3.3);
INSERT INTO table (`Text`, persons.`Person`) VALUES ('F1', ''), ('F2', '[{"id":"ou_<open_user_id>"}]');
Update table set `Select`='Y' WHERE record_id = 'XX';
Update table set `Select`='N' WHERE `Person` = '<person name>';
DELETE FROM table WHERE record_id = 'XX';
DELETE FROM table WHERE `person` = 'XX';
```

More examples, see [driver_test.go](driver/driver_test.go)

**特殊用法**:

- `show create view`: instead of 'show views'，use `show create view` show a view meta
- `create view kanban.{view_name} as select * from table`: when creating a view，`kanban` is the ViewType for view，more
  about ViewType: [model](doc/const.md) `ViewType`。
- "persons.\`person\`": a special type for person fieldType

**Special type**:
More about FieldType [model](doc/model.md)`FieldType`

- `persons`: json string， [model](doc/model.md)`RecordPerson`
- `url`: json string， [model](doc/model.md)`RecordUrl`
- `attachments`: json string， [model](doc/model.md)`RecordAttachments`
- `options`: json string， [model](doc/model.md)`RecordOptions`

example:

```
  url := `{"link":"https://www.google.com","text":"Google"}`
  options := `["option_one", "option_two"]`
  persons := `[{"id":"ou_fcb313360e8b813e8017771f6bbb9533"}]`
  attachments := `[{"file_token":"boxbcqtaK3s6cCsHPhzddAXVdhc"}]`
```

## use driver for code

```golang
package main

import (
	"database/sql"
	"fmt"
	"os"

	// load bitable driver
	_ "github.com/luw2007/bitable-mysql-driver/driver"
	"github.com/sirupsen/logrus"
)

var (
	appID     = os.Getenv("APP_ID")
	appSecret = os.Getenv("APP_SECRET")
	appToken  = os.Getenv("APP_TOKEN")
	dsn   = fmt.Sprintf("bitable://%s:%s@open.feishu.cn/%s", appID, appSecret, appToken)
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	db, err := sql.Open("bitable", dsn)
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		panic(err)
	}
	var tid, name, version string
	for rows.Next() {
		if err := rows.Scan(&tid, &name, &version); err != nil {
			panic(err)
		}
		fmt.Println(tid, name, version)
	}
}
```

$ go run main.go

```
tblTqyMTqUTFrDQc table1 4
tblebGSJc65Km9qG table2 11
tblssnegcbTL6pp2 table_create 5
```

## Bitable feature

### App

- [x] bitable meta

### table

- [x] list table
- [x] create table
- [x] delete table
- [x] batch delete table
- [ ] batch create table

### view

- [x] list view
- [x] add view
- [x] delete view

### record

- [x] list record
- [x] query record
- [x] add record
- [x] batch add record
- [x] update record
- [x] batch update record // use where condition，instead of `record_id`
- [x] delete record
- [x] multi delete record // use where condition，instead of `record_id`

### field

- [x] list field
- [x] add field
- [x] update field
- [x] delete field

## TODO:

- [ ] test suite，use [lark](https://github.com/chyroc/lark) mock api
- [ ] add to [usql](https://github.com/xo/usql)

## Thanks：

- [lark](https://github.com/chyroc/lark)
- [pingcap/parser](https://github.com/pingcap/parser)
