package notilib

import (
	"fmt"
)

type Retrialer interface {
	retry(content, guid string, index, numRetrials int)
}

type retrialer struct {
	msgCh chan message
}

func newRetrialer(msgChan chan message) (Retrialer, error) {
	if msgChan == nil {
		return nil, fmt.Errorf("msgChan can not be nil")
	}
	return &retrialer{
		msgCh: msgChan,
	}, nil
}

func (r *retrialer) retry(content, guid string, index, numRetrials int) {
	r.msgCh <- message{
		content:     content,
		guid:        guid,
		index:       index,
		numRetrials: numRetrials,
	}
}
