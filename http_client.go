package xecho

import (
	"net/http"
	"net/http/httputil"
	"time"

	newrelic "github.com/newrelic/go-agent"
)

type loggingTransport struct {
	inboundContext *Context
	isDebug        bool
	transport      http.RoundTripper
}

// Wraps the outbound request round trip with logging and metrics
func (t *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	segment := newrelic.StartExternalSegment(t.inboundContext.NewRelicTx, r)
	logger := t.inboundContext.Logger().(*Logger)

	if err := debugDumpRequest(r, logger, t.isDebug); err != nil {
		return nil, err
	}

	startTime := time.Now()

	res, err := t.transport.RoundTrip(r)

	segment.Response = res
	_ = segment.End()

	if err != nil {
		logger.Errorf("Failed to get response in outbound request: %s %s", r.Method, r.URL.String())
		return nil, err
	}

	reqTime := time.Now().Sub(startTime)

	logger.Infof("Outgoing request: %s %s %d (%fs)", r.Method, r.URL.String(), res.StatusCode, reqTime.Seconds())

	if err := debugDumpResponse(res, logger, t.isDebug); err != nil {
		return nil, err
	}

	return res, nil
}

func NewHttpClient(
	context *Context,
	client *http.Client,
) *http.Client {
	// wrap transport
	loggingTransport := &loggingTransport{
		inboundContext: context,
		isDebug:        context.IsDebug,
		transport:      client.Transport,
	}
	client.Transport = loggingTransport
	return client
}

func debugDumpRequest(r *http.Request, logger *Logger, isDebug bool) error {
	if !isDebug {
		return nil
	}

	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}

	logger.Debugf("%s", string(reqDump))
	return nil
}

func debugDumpResponse(res *http.Response, logger *Logger, isDebug bool) error {
	if !isDebug {
		return nil
	}

	resDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}

	logger.Debugf("%s", string(resDump))
	return nil
}
