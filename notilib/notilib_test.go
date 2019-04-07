package notilib

import (
	"net/http"
	"testing"
	"time"
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

func TestNotify(t *testing.T) {

	tt := []struct {
		name                  string
		url                   string
		conf                  *Config
		messages              []string
		expectedChannelLength int
		client                *http.Client
		errMsg                string
	}{
		{"Positive TC: default config", "http://localhost/api", nil, []string{"abc", "hello world", "hola mundo", "hallo Welt"}, 4, http.DefaultClient, ""},
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
			guid, err := notifier.Notify(tc.messages)
			if err != nil {
				t.Errorf("errors found when notifiying: %v", err)
				foundError = true
			}
			if guid == "" {
				t.Errorf("empty GUID")
				foundError = true
			}

			// give some time to insert the messages into the channel
			time.Sleep(1 * time.Second)

			if !foundError {
				if notifier == nil {
					t.Errorf("notifier is nil")
				}
				chLength := notifier.GetMessageChannelLength()
				if chLength != tc.expectedChannelLength {
					t.Errorf("unexpected number of items in message channel: expected=%d; got=%d", tc.expectedChannelLength, chLength)
				}
			}
		})
	}
}

func TestRetry(t *testing.T) {

	tt := []struct {
		name                  string
		url                   string
		conf                  *Config
		expectedChannelLength int
		client                *http.Client
		errMsg                string
	}{
		{"Positive TC: default config", "http://localhost/api", nil, 1, http.DefaultClient, ""},
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
			notifier.Retry("abc", "1111-2222-3333-4444", 5, 0)

			// give some time to insert the messages into the channel
			time.Sleep(1 * time.Second)

			if !foundError {
				if notifier == nil {
					t.Errorf("notifier is nil")
				}
				chLength := notifier.GetMessageChannelLength()
				if chLength != tc.expectedChannelLength {
					t.Errorf("unexpected number of items in message channel: expected=%d; got=%d", tc.expectedChannelLength, chLength)
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
