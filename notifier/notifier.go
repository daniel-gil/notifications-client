package notifier

import (
	"fmt"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Notifier interface {
	Notify(messages []string) (string, error)
	Retry(msg, guid string, index, numRetrials int)
	GetErrorChannel() chan NError
}

type notifier struct {
	url   string
	msgCh chan message
	errCh chan NError
}

const maxChCap = 10
const maxErrChCap = 100

func New(url string) Notifier {
	notifier := &notifier{
		url:   url,
		msgCh: make(chan message, maxChCap),
		errCh: make(chan NError, maxErrChCap),
	}
	go notifier.messageHandler()
	return notifier
}

func init() {
	// logrus configuration
	initLogger()
}

func initLogger() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func (n *notifier) messageHandler() {
	log.Debug("message handler started...")
	for {
		select {
		case msg := <-n.msgCh:
			log.Debugf("Handling new message: %s", msg.content)
			err := n.send(msg)
			if err != nil {
				n.errCh <- NError{
					GUID:        msg.guid,
					Index:       msg.index,
					Error:       err.Error(),
					Message:     msg.content,
					NumRetrials: msg.numRetrials + 1,
				}
			}
			log.Debugf("message sent")
		}
	}
}

func (n *notifier) Notify(messages []string) (string, error) {
	log.Debugf("queuing new messages: %s", messages)
	guid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("unable to create an GUID: %v", err)
	}

	// queueing messages into the channel to be later dispatched
	for idx, msg := range messages {
		n.msgCh <- message{
			content:     msg,
			guid:        guid.String(),
			index:       idx,
			numRetrials: 0,
		}
	}

	return guid.String(), nil
}

func (n *notifier) Retry(msg, guid string, index, numRetrials int) {
	n.msgCh <- message{
		content:     msg,
		guid:        guid,
		index:       index,
		numRetrials: numRetrials,
	}
}

func (n *notifier) GetErrorChannel() chan NError {
	return n.errCh
}

func (n *notifier) send(msg message) error {
	body := strings.NewReader(msg.content)
	req, err := http.NewRequest("POST", n.url, body)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send the request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP Status: %s", resp.Status)
	}
	return nil
}
