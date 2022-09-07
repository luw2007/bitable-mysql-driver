package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

var (
	appID     = os.Getenv("ONLINE_APP_ID")
	appSecret = os.Getenv("ONLINE_APP_SECRET")

	sheetTokenFlag = flag.String("s", "", "sheet token")
	VerboseFlag    = flag.Bool("v", false, "verbose")
)

func main() {
	flag.Parse()
	domain := "https://open.feishu.cn"
	s := lark.NewSheetsClient(appID, appSecret, domain, "trace")
	sheets, err := s.GetAPP(context.Background(), *sheetTokenFlag)
	if err != nil {
		panic(err)
	}
	for _, sheet := range sheets {
		if sheet.BlockInfo != nil && sheet.BlockInfo.BlockType == "BITABLE_BLOCK" {
			bitables := strings.SplitN(sheet.BlockInfo.BlockToken, "_", 2)
			fmt.Println(sheet.SheetID, sheet.Title, bitables[0], bitables[1])
		}
		if *VerboseFlag {
			b, _ := json.Marshal(sheet)
			fmt.Println(sheet.SheetID, string(b))
		}
	}
}
