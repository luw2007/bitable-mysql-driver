package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/readline"
	"github.com/sirupsen/logrus"

	_ "github.com/luw2007/bitable-mysql-driver/driver"
)

var (
	debugFlag = flag.Bool("d", false, "open debug log")
	jsonFlag  = flag.Bool("j", false, "json line output")
)

type Handler struct {
	Conn *sql.DB
}

func NewHandler(dsn string) *Handler {
	db, err := sql.Open("bitable", dsn)
	if err != nil {
		logrus.WithError(err).Fatal("connect open api error")
		panic(err)
	}
	return &Handler{Conn: db}
}

func (h Handler) SQL(sql string) {
	rows, err := h.Conn.Query(sql)
	if err != nil {
		fmt.Println("query error:", err)
		return
	}
	if rows == nil {
		return
	}
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("columns error:", err, ". Ignore when insert or update")
		return
	}
	fields := make([][]byte, len(columns))
	scans := make([]interface{}, len(columns))
	for i := range fields {
		scans[i] = &fields[i]
	}
	for rows.Next() {
		err := rows.Scan(scans...)
		if err != nil {
			fmt.Println("scan error:", err)
			return
		}
		if *jsonFlag {
			v := make(map[string]string, len(fields))
			for i, field := range fields {
				v[columns[i]] = strings.ReplaceAll(string(field), `"`, `\"`)
			}
			b, _ := json.Marshal(v)
			fmt.Printf("%s\n", string(b))
		} else {
			for i, field := range fields {
				fmt.Printf("%s: %v\n", columns[i], string(field))
			}
			fmt.Println("")
		}
	}
}

func (h Handler) Start() {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		h.SQL(line)
	}
}

func main() {
	flag.Parse()

	if *debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.SetOutput(os.Stdout)

	if flag.NArg() == 0 {
		logrus.Fatal("need dsn like: bitable://<app_id>:<app_secret>@open.feishu.cn/<app_token>")
	}
	NewHandler(flag.Args()[0]).Start()
}
