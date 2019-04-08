package notilib

import (
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	tt := []struct {
		name        string
		msgChan     chan message
		content     string
		guid        string
		index       int
		numRetrials int
		errMsg      string
	}{
		{"Positive TC", make(chan message, 20), "hello world", "1234", 4, 1, ""},
		{"Nil message channel", nil, "", "1234", 4, 1, "msgChan can not be nil"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			retrialer, err := newRetrialer(tc.msgChan)

			if tc.msgChan != nil {
				if !checkError(tc.errMsg, err, t) {
					retrialer.retry(tc.content, tc.guid, tc.index, tc.numRetrials)

					// give some time to call send method
					time.Sleep(1 * time.Second)

					if len(tc.msgChan) != 1 {
						t.Errorf("unexpected number of elements in msg chan: expected 1; got: %d", len(tc.msgChan))
					}
					if !checkMessageChannelContent([]string{tc.content}, tc.msgChan) {
						t.Errorf("unexpected content of elements inside msg chan")
					}
				}
			}
		})
	}
}
