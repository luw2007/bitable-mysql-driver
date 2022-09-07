# 多维表格(bitable) MySQL 驱动

飞书开放平台提供了多维表格功能。将其封装成 mysql driver，不用学习语法也能快速接入。

[English README](./readme.md)

## 快速上手

```bash
go run cmd/bsql/main.go 'bitable://cli_a14eda43cb7ad013:l5zyi***********************16Y0@open.feishu.cn/bascnQIrLs6MrhIvftGsdYJgRFd'
> show tables;
id: tblZAuXVNjvK3PE6
name: 测试读取-请勿修改
revision: 49

id: tbl9GhbLx2AyQTTd
name: 测试写入-请勿修改
revision: 132
...

> show create view tblZAuXVNjvK3PE6;
id: vewzs11WK3
name: 总表
type: grid

> show columns from tblZAuXVNjvK3PE6;
Field: fldE2eIpUf
Type: 1
Comment: ID
Extra: null

field: fldqPGPpRN
type: 2
comment: 数字
extra: {"formatter":"0"}

field: fld7SniGf3
type: 3
comment: 单选
extra: {"options":[{"name":"是","id":"optid4Goz3"},{"name":"否","id":"optdpSmpZj"}]}

...

> select * from tblZAuXVNjvK3PE6;
RecordId: recrXXouR5
...

> insert into tblZAuXVNjvK3PE6 (`多行文本`, persons.`关注`) values ('F8429', '[{"id":"ou_fcb313360e8b813e8017771f6bbb9533"}]');
> select `多行文本` from tblZAuXVNjvK3PE6 where `多行文本`='F8429';
RecordId: recQTyAF85
Fields: {"多行文本":"F8429"}

> update tblZAuXVNjvK3PE6 set `单选`='是' where record_id='recQTyAF85';
> update tblZAuXVNjvK3PE6 set `单选`='否' where `单选`='是';
```

## 常见场景：

```
SHOW TABLES;
SHOW COLUMNS FROM table;
SHOW CREATE VIEW table;

# 查询数据
SELECT * FROM table limit 10; 
SELECT * FROM table WHERE `数字` >= 2 and `人员` in ('XX')
limit 10; SELECT * FROM table WHERE `日期` >= TODATE('2021-12-16');
SELECT * FROM table WHERE `数字` in (3, 1) order by `数字` desc limit 10;
SELECT * FROM table WHERE record_id = 'rec9eOiv5d'; 
SELECT * FROM table WHERE 单选 IS NOT NULL;
SELECT * FROM table WHERE 单选 IS NULL;

# 记录操作
INSERT INTO table (`数字`) VALUES (3), (3.0), (0.3), (3.3);
INSERT INTO table (`多行文本`, persons.`关注`) VALUES ('F1', ''), ('F2', '[{"id":"ou_xx"}]');
Update table set `单选`='是' WHERE record_id = 'XX';
Update table set `单选`='是' WHERE `关注` = 'XX';
DELETE FROM table WHERE record_id = 'XX';
DELETE FROM table WHERE `关注` = 'XX';

# 表操作
CREATE TABLE table
(
    `PSM`  text,
    `选项`   varchar(3) COMMENT '{"options":[{"name":"是"}]}',
    `多人协作` varchar(11) COMMENT '{"multiple":true}'
) COMMENT '默认看板';
CREATE VIEW kanban.看板视图 AS SELECT * FROM table;
DROP TABLE table;

# 字段操作
ALTER TABLE table ADD COLUMN `多行文本` varchar(1) COMMENT '{"multiple":true}';
ALTER TABLE table CHANGE COLUMN `测试修改` `日期` varchar(5);
ALTER TABLE table MODIFY COLUMN `测试修改` text;
ALTER TABLE table RENAME COLUMN `日期` TO `测试修改`;
ALTER TABLE table DROP COLUMN `日期`;
```

更多使用例子见：[driver_test.go](driver/driver_test.go)

**特殊用法**：

- `show create view`：没有 'show views'，借用`show create view`作为视图查询
- `create view kanban.{view_name} as select * from table` 创建视图时候，`kanban` 表示看板类型，更多类型参考：[model](doc/const.md) `ViewType`
  。
- "persons.\`关注\`"：部分字段类型特殊，persons 表示是"人员"类型，这里接收一个列表
- `dbname`=`{table_id}.{view_id}`：如果需要`view_id`，所以将`table_id`+`view_id` 作为表名。

**特殊类型**： 字段类型见：[model](doc/model.md)`FieldType`

- `persons`: json 结构，对应具体看 [model](doc/model.md)`RecordPerson`
- `url`: json 结构，具体看 [model](doc/model.md)`RecordUrl`
- `attachments`：json 结构，具体看 [model](doc/model.md)`RecordAttachments`
- `options`：json 结构，具体看 [model](doc/model.md)`RecordOptions`

例子如下：

```
  url := `{"link":"https://www.google.com","text":"Google"}`
  options := `["选项一", "选项二"]`
  persons := `[{"id":"ou_fcb313360e8b813e8017771f6bbb9533"}]`
  attachments := `[{"file_token":"boxbcqtaK3s6cCsHPhzddAXVdhc"}]`
```

## 在代码中使用

```golang
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/luw2007/bitable-mysql-driver/driver"
)

var (
	appID     = os.Getenv("APP_ID")
	appSecret = os.Getenv("APP_SECRET")
	appToken  = os.Getenv("APP_TOKEN")
	testDSN   = fmt.Sprintf("bitable://%s:%s@open.feishu.cn/%s", appID, appSecret, appToken)
)

func main() {
	db, err := sql.Open("bitable", testDSN)
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("show tables")
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

```
$ go run main.go
tblHTHTelIVwIeqN 单向关联 4
tblmoAekspwLGzEt 双向关联 11
tblZAuXVNjvK3PE6 测试读取-请勿修改 49
tbl9GhbLx2AyQTTd 测试写入-请勿修改 132
tblssnegcbTL6pp2 table_create 5

```

## 多维表格特性支持

### App

- [x] 获取多维表格元数据

### 数据表

- [x] 列出数据表
- [x] 新增数据表
- [x] 删除数据表
- [x] 删除多个数据表
- [ ] 新增多个数据表

### 视图

- [x] 列出视图
- [x] 新增视图
- [x] 删除视图

### 记录

- [x] 列出记录
- [x] 检索记录
- [x] 新增记录
- [x] 新增多条记录
- [x] 更新记录
- [x] 更新多条记录 // 根据条件更新，并非 record
- [x] 删除记录
- [x] 删除多条记录 // 根据条件更新，并非 record

### 字段

- [x] 列出字段
- [x] 新增字段
- [x] 更新字段
- [x] 删除字段

PS: 删除功能并非无法实现，主要是危险性比较高。

## TODO:

- [x] 基本特性
- [x] 交互模式的查询工具
- [x] stmt、rows 代码重构
- [x] 支持 user_access_token
- [ ] 测试用例集，使用 mockapi
- [ ] gorm
- [ ] 支持 usql

## 感谢：

- [lark](https://github.com/chyroc/lark)
- [pingcap/parser](https://github.com/pingcap/parser)
- [copier](https://github.com/jinzhu/copier)