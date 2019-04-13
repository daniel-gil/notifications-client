package notilib

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Listener interface {
	listen(ctx context.Context)
	flush(timeout time.Duration, quit chan<- bool)
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
func (l *requestHandler) listen(ctx context.Context) {
	// generate tickets each `rate` to avoid overwhelm the server
	tick := time.Tick(l.rate)

	for {
		select {
		case msg := <-l.msgChan:
			// here got a new message from the Message Channel
			<-tick
			// here got a ticket to process the message
			go l.sender.send(msg)
		case <-ctx.Done():
			log.Infof("listen: %v", ctx.Err())
			return
		}
	}
}

func (l requestHandler) flush(timeout time.Duration, quit chan<- bool) {
	log.Debugf("Listener: %d messages to be flushed", len(l.msgChan))

	// programming timeout
	t := time.After(timeout)
	go func(quit chan<- bool) {
		<-t
		log.Warn("timeout occurs flushing notifications")
		quit <- true
		return
	}(quit)

	numMessages := len(l.msgChan)
	for i := 1; i <= numMessages; i++ {
		log.Debugf("Listener: flushing #%d message", i)

		msg := <-l.msgChan
		l.sender.send(msg)
	}
	log.Infof("flushed %d messages", numMessages)
	quit <- true
	return
}
