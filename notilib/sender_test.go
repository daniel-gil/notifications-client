package notilib

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type MockDispatcher struct {
	dispatchMock func(*http.Request) (*http.Response, error)
}

func (m *MockDispatcher) dispatch(req *http.Request) (*http.Response, error) {
	return m.dispatchMock(req)
}

type contextMode int

const (
	contextDoneNotCalled contextMode = iota
	contextDoneCalledBeforeSend
	contextDoneCalledAfterSend
)

func TestSend(t *testing.T) {
	tt := []struct {
		name     string
		url      string
		ctxMode  contextMode
		testData string
	}{
		{"Positive TC", "http://localhost", contextDoneNotCalled, "body content"},
		{"Negative TC: context done before send", "http://localhost", contextDoneCalledBeforeSend, "body content"},
		{"Negative TC: context done after send", "http://localhost", contextDoneCalledAfterSend, "body content"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var cancel context.CancelFunc
			called := false
			mockDispatcher := &MockDispatcher{
				dispatchMock: func(req *http.Request) (*http.Response, error) {
					called = true

					defer req.Body.Close()
					bodyBytes, err := ioutil.ReadAll(req.Body)
					if err != nil {
						t.Errorf("unable to read the response body: %v", err)
					}
					bodyString := string(bodyBytes)
					if bodyString != tc.testData {
						t.Errorf("expected this test data: %v; got %v", tc.testData, bodyString)
					}
					if bodyString != tc.testData {
						t.Errorf("requet body mismatch. Expected \"%s\"; got: \"%s\"", tc.testData, bodyString)
					}
					return createHTTPResponse(req, tc.testData), nil
				},
			}
			errCh := make(chan NError, 10)
			sender := NewSender(tc.url, mockDispatcher, errCh)
			ctx := context.Background()

			if tc.ctxMode == contextDoneCalledBeforeSend {
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-7*time.Hour))
				cancel()
			}

			msg := getDummyMessage(tc.testData)
			go sender.send(msg)

			// give some time to call send method
			time.Sleep(2 * time.Second)

			if !called {
				t.Errorf("did not call client.dispatch")
			}
		})
	}
}

func createHTTPResponse(req *http.Request, body string) *http.Response {
	return &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Request:    req,
		StatusCode: http.StatusOK,
	}
}
