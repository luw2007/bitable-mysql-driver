package lark

import (
	"bytes"
	"context"
	"testing"

	larksdk "github.com/chyroc/lark"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLarkLoggerDropsSensitiveVerboseMessages(t *testing.T) {
	output := new(bytes.Buffer)
	logger := logrus.New()
	logger.SetOutput(output)
	logger.SetLevel(logrus.TraceLevel)
	bridge := larkLogger{logger: logger}

	bridge.Log(context.Background(), larksdk.LogLevelTrace, "request header=Bearer secret")
	bridge.Log(context.Background(), larksdk.LogLevelDebug, "response tenant_access_token=secret")
	bridge.Log(context.Background(), larksdk.LogLevelInfo, "request completed")

	require.NotContains(t, output.String(), "secret")
	require.Contains(t, output.String(), "request completed")
}

func TestLarkLogLevelClampsVerboseConfiguration(t *testing.T) {
	require.Equal(t, larksdk.LogLevelInfo, getLarkLogLevel("trace"))
	require.Equal(t, larksdk.LogLevelInfo, getLarkLogLevel("debug"))
	require.Equal(t, larksdk.LogLevelWarn, getLarkLogLevel("warn"))
}
