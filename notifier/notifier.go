package notifier

import (
	"fmt"
	"net/http"
	neturl "net/url"
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
	GetConfig() *Config
	StartService()
	GetMessageChannelLength() int
}

type notifier struct {
	conf   *Config
	client *http.Client
	url    string
	msgCh  chan message
	errCh  chan NError
}

const defaultMaxChCap = 1000
const defaultMaxErrChCap = 500
const defaultBurstLimit = 1000
const defaultNumMessagesPerSecond = 1000

// New creates a new object that implements Notifier interface
func New(url string, conf *Config) (Notifier, error) {

	err := checkURLFormat(url)
	if err != nil {
		return nil, err
	}

	// if no configuration is provided, generate the default configuration
	conf = getConfiguration(conf)
	log.Debugf("Notifier configuration: \n%v\n", conf)

	notifier := &notifier{
		conf:   conf,
		url:    url,
		client: &http.Client{},
		msgCh:  make(chan message, conf.MaxChCap),
		errCh:  make(chan NError, conf.MaxErrChCap),
	}
	return notifier, nil
}

func (n *notifier) GetConfig() *Config {
	return n.conf
}

func (n *notifier) StartService() {
	go func() {
		rate := time.Second / time.Duration(n.conf.NumMessagesPerSecond)
		tick := time.NewTicker(rate)
		defer tick.Stop()
		throttle := make(chan time.Time, n.conf.BurstLimit)
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
	}()
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

func (n *notifier) GetMessageChannelLength() int {
	return len(n.msgCh)
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

	// defer the close operation of the response body to avoid a resource leak
	defer resp.Body.Close()

	// check if the response is a successful HTTP code: 200 OK or 201 Created
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected HTTP Status: %s", resp.Status)
	}
	return nil
}

func checkURLFormat(url string) error {
	if url == "" {
		return fmt.Errorf("empty URL")
	}
	_, err := neturl.ParseRequestURI(url)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	return nil
}
