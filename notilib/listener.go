package notilib

import (
	"fmt"
	"time"
)

type Listener interface {
	listen()
}

type requestHandler struct {
	rate       time.Duration
	burstLimit int
	msgChan    chan message
	sender     sender
}

func newListener(r time.Duration, b int, ch chan message, s sender) (Listener, error) {
	if ch == nil {
		return nil, fmt.Errorf("message channel can not be nil")
	}
	if s == nil {
		return nil, fmt.Errorf("sender can not be nil")
	}
	return &requestHandler{
		rate:       r,
		burstLimit: b,
		msgChan:    ch,
		sender:     s,
	}, nil
}

// listen waits for receiving new notifications from the request channel and processes them
func (l *requestHandler) listen() {
	tick := time.NewTicker(l.rate)
	defer tick.Stop()
	throttle := make(chan time.Time, l.burstLimit)
	go func() {
		for t := range tick.C {
			select {
			case throttle <- t:
			default:
			}
		}
	}()
	for msg := range l.msgChan {
		<-throttle
		go l.sender.send(msg)
	}
}
