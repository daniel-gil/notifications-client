package notilib

import (
	"testing"
	"time"
)

type MockSender struct {
	sendMock func(msg message)
}

func (m *MockSender) send(msg message) {
	m.sendMock(msg)
}

func TestListen(t *testing.T) {
	tt := []struct {
		name            string
		reqChanCapacity int
		senderNil       bool
		testData        string
		errMsg          string
	}{
		{"Positive TC", 10, false, "body content", ""},
		{"Negative TC: nil request channel", -1, false, "body content", "message channel can not be nil"},
		{"Negative TC: nil sender", 1, true, "body content", "sender can not be nil"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			var channel chan message
			if tc.reqChanCapacity != -1 {
				channel = make(chan message, tc.reqChanCapacity)

				// add a dummy msg to be processed in the listen() function
				msg := getDummyMessage(tc.testData)
				channel <- msg
			}

			called := false
			var mockSender sender
			if !tc.senderNil {
				mockSender = &MockSender{
					sendMock: func(msg message) {
						called = true
					},
				}
			}

			listener, err := NewListener(1*time.Second, 10, channel, mockSender)
			if !checkError(tc.errMsg, err, t) {
				go listener.listen()

				// give some time to call send method
				time.Sleep(2 * time.Second)

				if !called {
					t.Fatalf("did not call fulfill")
				}
			}
		})
	}
}

func getDummyMessage(content string) message {
	return message{
		content:     content,
		guid:        "111-222-333-444",
		index:       3,
		numRetrials: 0,
	}
}
