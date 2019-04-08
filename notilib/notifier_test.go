package notilib

import (
	"testing"
	"time"
)

func TestNotify(t *testing.T) {
	tt := []struct {
		name     string
		msgChan  chan message
		messages []string
		errMsg   string
	}{
		{"Positive TC", make(chan message, 20), []string{"abc", "zzzz", "hello world"}, ""},
		{"Nil message channel", nil, nil, "msgChan can not be nil"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			notifier, err := newNotifier(tc.msgChan)

			if tc.msgChan != nil {
				if !checkError(tc.errMsg, err, t) {
					notifier.notify(tc.messages)

					// give some time to call send method
					time.Sleep(1 * time.Second)

					if len(tc.messages) != len(tc.msgChan) {
						t.Errorf("unexpected number of elements in msg chan: expected %d; got: %d", len(tc.messages), len(tc.msgChan))
					}
					if !checkMessageChannelContent(tc.messages, tc.msgChan) {
						t.Errorf("unexpected content of elements inside msg chan")
					}
				}
			}
		})
	}
}

func checkMessageChannelContent(messages []string, ch chan message) bool {
	chLen := len(ch)
	for i := 0; i < chLen; i++ {
		msg := <-ch

		idx := containsInSlice(msg.content, messages)
		if idx == -1 {
			// not found
			return false
		}

		// found, remove from messages slice
		messages = removeFromSlice(messages, idx)
	}
	if len(messages) == 0 {
		// we found all items from messages inside the channel
		return true
	}
	return false
}

func containsInSlice(value string, messages []string) int {
	for idx, msg := range messages {
		if msg == value {
			return idx
		}
	}
	return -1
}

func removeFromSlice(slice []string, idx int) []string {
	return append(slice[:idx], slice[idx+1:]...)
}
