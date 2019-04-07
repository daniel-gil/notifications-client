package notifier

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Notifier is the interface for sending notifications
type Notifier interface {
	Notify(messages []string) (string, error)
	Retry(msg, guid string, index, numRetrials int)
	GetErrorChannel() <-chan NError
}

type notifier struct {
	client *http.Client
	url    string
	msgCh  chan message
	errCh  chan NError
}

const prefix = "notifier"
const defaultMaxChCap = 1000
const defaultMaxErrChCap = 500
const defaultBurstLimit = 1000
const defaultNumMessagesPerSecond = 1000

// New creates a new object that implements Notifier interface
func New(url string, conf *Config) Notifier {

	// if no configuration is provided, generate the default configuration
	conf = getConfiguration(conf)
	log.Debugf("Notifier configuration: \n%v\n", conf)

	notifier := &notifier{
		url:    url,
		client: &http.Client{},
		msgCh:  make(chan message, conf.MaxChCap),
		errCh:  make(chan NError, conf.MaxErrChCap),
	}
	go notifier.startService(conf.NumMessagesPerSecond, conf.BurstLimit)
	return notifier
}

func (n *notifier) startService(numMessagesPerSecond, burstLimit int) {
	rate := time.Second / time.Duration(numMessagesPerSecond)
	tick := time.NewTicker(rate)
	defer tick.Stop()
	throttle := make(chan time.Time, burstLimit)
	go func() {
		for t := range tick.C {
			select {
			case throttle <- t:
			default:
			}
		}
	}()
	for msg := range n.msgCh {
		<-throttle
		go n.dispatchMessage(msg)
	}
}

func init() {
	// logrus configuration
	initLogger()
}

func getConfiguration(conf *Config) *Config {
	if conf == nil {
		conf = buildDefaultConfiguration()
	}
	return conf
}

func buildDefaultConfiguration() *Config {
	return &Config{
		BurstLimit:           defaultBurstLimit,
		NumMessagesPerSecond: defaultNumMessagesPerSecond,
		MaxChCap:             defaultMaxChCap,
		MaxErrChCap:          defaultMaxErrChCap,
	}
}

func initLogger() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func (n *notifier) Notify(messages []string) (string, error) {
	log.Debugf("queuing new messages: %s", messages)
	guid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("unable to create an GUID: %v", err)
	}

	// queueing messages into the channel to be later dispatched
	go func() {
		for idx, msg := range messages {
			// just queue those messages with content
			if len(msg) > 0 {
				n.msgCh <- message{
					content:     msg,
					guid:        guid.String(),
					index:       idx,
					numRetrials: 0,
				}
			}
		}
	}()
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

func (n *notifier) GetErrorChannel() <-chan NError { // returns receive-only channel of NError
	return n.errCh
}

func (n *notifier) dispatchMessage(msg message) {
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
}

func (n *notifier) send(msg message) error {
	body := strings.NewReader(msg.content)
	req, err := http.NewRequest("POST", n.url, body)
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send the request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected HTTP Status: %s", resp.Status)
	}
	return nil
}
