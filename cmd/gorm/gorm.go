package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/luw2007/bitable-mysql-driver/driver"
)

type Leads struct {
	RecordID          string `gorm:"column:record_id; type:text"` // 特殊列名
	Text              string `gorm:"column:主键ID;  type:text"`
	Number            string `gorm:"column:数字;  type:text"`
	Select            string `gorm:"column:单选; type:text"`
	MultipleSelect    string `gorm:"column:多选;  type:text"`
	Date              string `gorm:"column:日期;  type:text"`
	Checkbox          string `gorm:"column:复选框;  type:text"`
	Person            string `gorm:"column:人员;  type:text"`
	Link              string `gorm:"column:超链接;  type:text"`
	Attachment        string `gorm:"column:附件;  type:text"`
	OneWayAssociation string `gorm:"column:单向关联;  type:text"`
	ReferenceLookup   string `gorm:"column:引用查找;  type:text"`
	Formula           string `gorm:"column:公式;  type:text"`
	TwoWayAssociation string `gorm:"column:双向关联;  type:text"`
	CreateTime        string `gorm:"column:创建时间;  type:text"`
	UpdateTime        string `gorm:"column:最后更新时间;  type:text"`
	Founder           string `gorm:"column:创建人;  type:text"`
	Modifier          string `gorm:"column:修改人;  type:text"`
}

func (Leads) TableName() string { return testTable1 }

var (
	appID      = os.Getenv("APP_ID")
	appSecret  = os.Getenv("APP_SECRET")
	appToken   = os.Getenv("APP_TOKEN")
	testTable1 = os.Getenv("TABLE_1") + "5"

	dsn = fmt.Sprintf("bitable://%s:%s@open.feishu.cn/%s?log_level=trace", appID, appSecret, appToken)
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	sqlDB, err := sql.Open("bitable", dsn)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(
		mysql.New(mysql.Config{
			Conn:                      sqlDB,
			SkipInitializeWithVersion: true,
		}),
		&gorm.Config{})
	// Read

	err = db.Migrator().CreateTable(Leads{})
	if err != nil {
		panic(err)
	}
	var leads Leads
	db.Where("数字=?", 1).Where("ID=?", "F1").First(&leads) // 根据整形主键查找
	fmt.Printf("leads %+v", leads)
}
