package driver

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/pingcap/parser"
	"github.com/sirupsen/logrus"

	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

// Driver bitable driver for implement database/sql/driver
type Driver struct {
}

func init() {
	logrus.Debug("Initializing bitable driver")
	sql.Register("bitable", &Driver{})
}

// Open for implement driver interface
func (driver *Driver) Open(name string) (driver.Conn, error) {
	u, err := url.Parse(name)
	if err != nil {
		return nil, fmt.Errorf("parse dsn error: %w", err)
	}
	if u.Scheme != BitableSchema {
		return nil, fmt.Errorf("[bitable driver]  unsupported scheme %s", u.Scheme)
	}

	logrus.Debug("[bitable driver]  exec open")

	domain := fmt.Sprintf("https://%s", u.Hostname())

	appID := u.User.Username()
	appSecret, ok := u.User.Password()
	if len(appID) == 0 || !ok {
		return nil, errors.New("[bitable driver]  needed username and password")
	}
	logLevel := "info"
	querys := u.Query()
	if l := querys.Get("log_level"); l != "" {
		logLevel = l
	}
	if debug := querys.Get("debug"); debug != "" {
		logLevel = "trace"
	}
	timeoutStr := querys.Get("timeout")
	if timeoutStr == "" {
		timeoutStr = "5s"
	}
	ts, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("parse timeout err, timeout:%s, %w", timeoutStr, err)
	}

	logrus.Debugf("[bitable driver]  log level %v", logLevel)

	conn := &Conn{
		BiTable:   lark.NewLarkClient(appID, appSecret, domain, logLevel, ts),
		AppID:     appID,
		AppSecret: appSecret,
		parser:    parser.New(),
		AppToken:  u.Path[1:],
	}
	return conn, nil
}
