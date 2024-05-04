package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/luw2007/bitable-mysql-driver/driver"
	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

var (
	appID     = os.Getenv("APP_ID")
	appSecret = os.Getenv("APP_SECRET")

	SRCAppToken = flag.String("f", "", "from bitable app token")
	SRCTable    = flag.String("fd", "", "from bitable table name, optional")
	DSTAppToken = flag.String("t", "", "to bitable app token")
)

func main() {
	flag.Parse()
	domain := "https://open.feishu.cn"
	if appID == "" || appSecret == "" || *SRCAppToken == "" || *DSTAppToken == "" {
		flag.CommandLine.Usage()
		return
	}
	b := lark.NewLarkClient(appID, appSecret, domain, "trace", 10*time.Second)
	ctx := context.Background()

	tables := map[string]string{}
	page, err := b.ListALLTable(ctx, *SRCAppToken)
	if err != nil {
		panic(err)
	}

	for _, t := range page.Items {
		table := t.(*lark.Table)
		if *SRCTable == "" || *SRCTable == table.Name {
			tables[table.Name] = table.TableID
		}
	}

	for name, tableID := range tables {
		fields, err := b.ListAllFields(ctx, *SRCAppToken, tableID)
		if err != nil {
			panic(err)
		}
		if *SRCAppToken == *DSTAppToken {
			name = name + "_" + tableID
		}
		tableID, err := b.CreateTable(ctx, *DSTAppToken, name)
		if err != nil {
			panic(err)
		}

		newFields, err := b.ListAllFields(ctx, *SRCAppToken, tableID)
		if err != nil {
			panic(err)
		}
		defualtFieldID := newFields.Items[0].(*lark.Field).FieldID
		for i, f := range fields.Items {
			field := f.(*lark.Field)
			property := ""
			if field.Property != nil {
				b, _ := json.Marshal(field.Property)
				property = string(b)
			}
			fieldType := field.Type
			// 特殊处理关联的情况
			if fieldType == int64(driver.FieldTypeOneWayAssociation) ||
				fieldType == int64(driver.FieldTypeReferenceLookup) ||
				fieldType == int64(driver.FieldTypeTwoWayAssociation) {
				fieldType = int64(driver.FieldTypeText)
			}
			if i == 0 {
				_, err := b.UpdateField(ctx, *DSTAppToken, tableID, defualtFieldID, field.FieldName, fieldType, property)
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			_, err := b.AddField(ctx, *DSTAppToken, tableID, field.FieldName, fieldType, property)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
