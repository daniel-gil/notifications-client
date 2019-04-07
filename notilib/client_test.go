package notilib

import (
	"log"
	"net/http"
	"testing"
)

type MockHTTPClient struct {
	DoMock func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoMock(req)
}

func TestDispatch(t *testing.T) {
	tt := []struct {
		name string
	}{
		{"Positive TC"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			mockHTTPClient := &MockHTTPClient{
				DoMock: func(req *http.Request) (*http.Response, error) {
					called = true
					return &http.Response{}, nil
				},
			}

			sender := NewClientHandler(mockHTTPClient)
			hreq := createHTTPRequest()
			res, err := sender.dispatch(hreq)
			if err != nil {
				t.Errorf("response with errors: %v", err)
			}
			if res == nil {
				t.Errorf("empty response")
			}
			if !called {
				t.Fatalf("did not call client.Do")
			}
		})
	}
}

func createHTTPRequest() *http.Request {
	url := "http://date.jsontest.com/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}
	return req
}
