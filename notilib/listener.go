package notilib

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Listener interface {
	listen(ctx context.Context)
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
	tick := time.NewTicker(l.rate)
	defer tick.Stop()
	throttle := make(chan time.Time, l.burstLimit)
	go func() {
		for t := range tick.C {
			select {
			case throttle <- t:
			case <-ctx.Done():
				log.Infof("Tick generator: %v", ctx.Err())
				return
			}
		}
	}()

	select {
	case msg := <-l.msgChan:
		<-throttle
		go l.sender.send(msg)
	case <-ctx.Done():
		log.Infof("listen: %v", ctx.Err())
		return
	}
}
