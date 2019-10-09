package xecho

import (
	"github.com/sirupsen/logrus"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestHttpClient(t *testing.T) {
	cli := NewHttpClient(&Context{logger: &Logger{logrus.New().WithFields(logrus.Fields{})}, IsDebug: true}, &http.Client{Transport: &http.Transport{}})
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

func TestHttpClientPassthrough(t *testing.T) {

	timeout := time.Second * 23
	someClient := &http.Client{
		Timeout: timeout,
	}
	cli := NewHttpClient(&Context{}, someClient)
	assert.Equal(t, timeout, cli.Timeout)
}
