package notilib

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Notifier interface {
	// Notifies notifications to the configured URL
	Notify(messages []string) (string, error)
}

type Retrialer interface {
	// Retries sending failed notifications
	Retry(msg, guid string, index, numRetrials int)
}

type Notilib interface {
	Notifier
	Retrialer

	// Start the service that reads from the Message Channel and send them to the URL
	StartService()

	// Retrives the number of elements in the Message Channel pending to be notified
	GetMessageChannelLength() int

	// Retrieves the Error Channel for reading operations (to be able to handle those errors)
	GetErrorChannel() <-chan NError
}

type notilib struct {
	conf     *Config
	client   *http.Client
	url      string
	msgCh    chan message
	errCh    chan NError
	listener listener
}

const defaultMaxChCap = 1000
const defaultMaxErrChCap = 500
const defaultBurstLimit = 1000
const defaultNumMessagesPerSecond = 1000
const defaultReqChanCapacity = 100

// New creates a new object that implements Notifier interface
func New(url string, client *http.Client, conf *Config) (Notilib, error) {

	err := checkURLFormat(url)
	if err != nil {
		return nil, err
	}

	// if no configuration is provided, generate the default configuration
	conf = getConfiguration(conf)
	log.Debugf("Notifier configuration: \n%v\n", conf)

	if client == nil {
		client = http.DefaultClient
	}

	msgChan := make(chan message, conf.MaxChCap)
	errCh := make(chan NError, conf.MaxErrChCap)

	// build services to be injected
	clientHandler := newClientHandler(client)
	sender := newSender(url, clientHandler, errCh)
	rate := time.Second / time.Duration(conf.NumMessagesPerSecond)
	listener, err := newListener(rate, conf.BurstLimit, msgChan, sender)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize notilib: %v", err)
	}

	notilib := &notilib{
		listener: listener,
		conf:     conf,
		url:      url,
		client:   &http.Client{},
		msgCh:    msgChan,
		errCh:    errCh,
	}

	return notilib, nil
}

func (n *notilib) Notify(messages []string) (string, error) {
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

func (n *notilib) Retry(msg, guid string, index, numRetrials int) {
	n.msgCh <- message{
		content:     msg,
		guid:        guid,
		index:       index,
		numRetrials: numRetrials,
	}
}

func (n *notilib) StartService() {
	go n.listener.listen()
}

func (n *notilib) GetMessageChannelLength() int {
	return len(n.msgCh)
}

func (n *notilib) GetErrorChannel() <-chan NError { // returns receive-only channel of NError
	return n.errCh
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
		ReqChanCapacity:      defaultReqChanCapacity,
	}
}
