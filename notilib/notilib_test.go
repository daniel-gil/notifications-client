package notilib

import (
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	conf := &Config{
		MaxChCap:             1,
		MaxErrChCap:          2,
		BurstLimit:           3,
		NumMessagesPerSecond: 4,
	}

	tt := []struct {
		name   string
		url    string
		conf   *Config
		client *http.Client
		errMsg string
	}{
		{"Positive TC: default config", "http://localhost/api", nil, http.DefaultClient, ""},
		{"Positive TC: custom config", "http://localhost/api", conf, http.DefaultClient, ""},
		{"Missing URL", "", nil, http.DefaultClient, "empty URL"},
		{"Invalid URL", "http/abc", nil, http.DefaultClient, "invalid URL"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var foundError = false

			notifier, err := New(tc.url, tc.client, tc.conf)
			if err != nil {
				if checkError(tc.errMsg, err, t) {
					foundError = true
				}
			}

			if !foundError {
				if notifier == nil {
					t.Errorf("notifier is nil")
				}
			}
		})
	}
}

func checkError(errMsg string, err error, t *testing.T) bool {
	if err != nil {
		if errMsg == "" {
			// here the testcase didn't expect any error
			t.Errorf("unexpected error: %v", err)
		} else if errMsg != err.Error() {
			// here the testcase expected another error than the received
			t.Errorf("expected error message: %v; got: %v", errMsg, err.Error())
		}
		return true
	}
	return false
}
