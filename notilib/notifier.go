package notilib

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Notifier interface {
	notify(messages []string) (string, error)
}

type notifier struct {
	msgCh chan message
}

func newNotifier(msgChan chan message) (Notifier, error) {
	if msgChan == nil {
		return nil, fmt.Errorf("msgChan can not be nil")
	}
	return &notifier{
		msgCh: msgChan,
	}, nil
}

func (n *notifier) notify(messages []string) (string, error) {
	log.Debugf("queuing new messages: %s", messages)
	guid, err := n.newGUID()
	if err != nil {
		return "", err
	}

	// queueing messages into the channel to be later dispatched
	go func(messages []string) {
		for idx, msg := range messages {
			// just queue those messages with content
			if len(msg) > 0 {
				n.msgCh <- message{
					content:     msg,
					guid:        guid,
					index:       idx,
					numRetrials: 0,
				}
			}
			log.Debugf("message[%d] added, content=%s", idx, msg)
		}
		log.Debugf("%d messages inserted into the msgCh", len(messages))
	}(messages)
	return guid, nil
}

func (n *notifier) newGUID() (string, error) {
	guid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("unable to create an GUID: %v", err)
	}
	return guid.String(), nil
}
