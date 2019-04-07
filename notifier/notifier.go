package notifier

import (
	"fmt"
	"net/http"
	"strings"

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
	url   string
	msgCh chan message
	errCh chan NError
}

const prefix = "notifier"
const defaultNumWorkers = 5
const defaultMaxChCap = 10
const defaultMaxErrChCap = 100

// New creates a new object that implements Notifier interface
func New(url string, conf *Config) Notifier {

	// if no configuration is provided, generate the default configuration
	conf = getConfiguration(conf)
	log.Debugf("Notifier configuration: \n%v\n", conf)

	notifier := &notifier{
		url:   url,
		msgCh: make(chan message, conf.MaxChCap),
		errCh: make(chan NError, conf.MaxErrChCap),
	}

	// create workers to handle the messages that arrive to the channel
	notifier.startWorkers(conf.NumWorkers)

	return notifier
}

func init() {
	// logrus configuration
	initLogger()
}

func getConfiguration(conf *Config) *Config {
	if conf == nil {
		conf = buildDefaultConfiguration()
	}
	if conf.NumWorkers <= 0 {
		// we need at least 1 worker to process notifications
		conf.NumWorkers = 1
	}
	return conf
}

func buildDefaultConfiguration() *Config {
	return &Config{
		NumWorkers:  defaultNumWorkers,
		MaxChCap:    defaultMaxChCap,
		MaxErrChCap: defaultMaxErrChCap,
	}
}

func initLogger() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func (n *notifier) startWorkers(numWorkers int) {
	for w := 1; w <= numWorkers; w++ {
		go n.worker(w, n.msgCh)
	}
}

func (n *notifier) worker(id int, ch <-chan message) {
	for msg := range ch {
		log.Debugf("Worker[%v] processing message: [%s]", id, msg.content)
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
}

func (n *notifier) Notify(messages []string) (string, error) {
	log.Debugf("queuing new messages: %s", messages)
	guid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("unable to create an GUID: %v", err)
	}

	// queueing messages into the channel to be later dispatched
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

func (n *notifier) send(msg message) error {
	body := strings.NewReader(msg.content)
	req, err := http.NewRequest("POST", n.url, body)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send the request: %v", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected HTTP Status: %s", resp.Status)
	}
	return nil
}
