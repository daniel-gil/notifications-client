package notilib

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
	// update the number of retrials
	retrials := numRetrials + 1

	r.msgCh <- message{
		content:     content,
		guid:        guid,
		index:       index,
		numRetrials: retrials,
	}
	log.Warnf("Retrial[%v]: { GUID : \"%s\", Index : %v, Content : \"%s\" }", retrials, guid, index, content)
}
