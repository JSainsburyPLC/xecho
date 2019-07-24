package xecho

import (
	"github.com/sirupsen/logrus"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestHttpClient(t *testing.T) {
	cli := NewHttpClient(&Context{logger: &Logger{logrus.New().WithFields(logrus.Fields{})}}, true)
	apitest.NewMock().
		HttpClient(cli).
		Get("http://example.com/message").
		RespondWith().
		Status(http.StatusAccepted).
		EndStandalone()

	resp, err := cli.Get("http://example.com/message")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}
