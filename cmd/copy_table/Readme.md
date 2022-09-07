# 多维表格 copy 表头

```shell
Usage copy_table
  -d    open debug log
  -f string
        from bitable app token
  -fd string
        from bitable table name
  -t string
        to bitable app token

```
APP_ID=<app_id> APP_SECRET=<app_secret> go run cmd/copy_table/main.go -f <from_app_token> -t <to_app_token>  -fd <from_table_name>

app_id：开放平台应用ID
app_secret：开放平台应用密钥
from_app_token: 复制来源
to_app_token: 复制到
from_table_name: 表格命令，不传则是同步全部表头

PS: 暂时不支持: 单向关联、引用查找 、双向关联