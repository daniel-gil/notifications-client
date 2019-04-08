package notilib

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

const defaultMsgChCap = 1000
const defaultErrChCap = 500
const defaultBurstLimit = 1000
const defaultNumMessagesPerSecond = 1000
const defaultLogLevel = log.InfoLevel

// Notilib interface exposes the public methods of the library
type Notilib interface {
	// Listen start the service that reads from the Message Channel and send them to the URL
	Listen()

	// Notify queues the messages into the Message Channel
	Notify(messages []string) (string, error)

	// Retry queue a message structure into the Message Channel
	Retry(msg, guid string, index, numRetrials int)

	// Retrieves the receive-only Error Channel for reading operations (to be able to handle those errors)
	GetErrorChannel() <-chan NError
}

type notilib struct {
	errCh     chan NError
	listener  Listener
	notifier  Notifier
	retrialer Retrialer
}

// New creates a new object that implements Notilib interface
func New(url string, client *http.Client, conf *Config) (Notilib, error) {

	// if no configuration is provided, build a default configuration
	conf = getConfiguration(conf)

	// logrus configuration
	initLogger(conf.LogLevel)
	log.Debugf("Notilib configuration: \n%v\n", conf)

	// validate the URL format
	err := checkURLFormat(url)
	if err != nil {
		return nil, err
	}

	// create channels
	msgChan := make(chan message, conf.MsgChanCap)
	errCh := make(chan NError, conf.ErrChanCap)

	// create a listener
	listener, err := buildListener(url, conf, client, msgChan, errCh)
	if err != nil {
		return nil, err
	}

	// create a notifier
	notifier, err := newNotifier(msgChan)
	if err != nil {
		return nil, err
	}

	// create a retrialer
	retrialer, err := newRetrialer(msgChan)
	if err != nil {
		return nil, err
	}

	notilib := &notilib{
		errCh:     errCh,
		listener:  listener,
		notifier:  notifier,
		retrialer: retrialer,
	}

	return notilib, nil
}

func buildListener(url string, conf *Config, client *http.Client, msgChan chan message, errCh chan NError) (Listener, error) {
	if client == nil {
		client = http.DefaultClient
	}
	clientHandler := newClientHandler(client)
	sender := newSender(url, clientHandler, errCh)
	rate := time.Second / time.Duration(conf.NumMessagesPerSecond)
	listener, err := newListener(rate, conf.BurstLimit, msgChan, sender)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize notilib: %v", err)
	}
	return listener, err
}

func (n *notilib) Notify(messages []string) (string, error) {
	return n.notifier.notify(messages)
}

func (n *notilib) Retry(msg, guid string, index, numRetrials int) {
	n.retrialer.retry(msg, guid, index, numRetrials)
}

func (n *notilib) Listen() {
	go n.listener.listen()
}

func (n *notilib) GetErrorChannel() <-chan NError {
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

func initLogger(logLevel log.Level) {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(logLevel)
}

func getConfiguration(conf *Config) *Config {
	if conf == nil {
		return DefaultConfig()
	}
	if conf.BurstLimit < 0 {
		conf.BurstLimit = defaultBurstLimit
	}
	if conf.MsgChanCap < 0 {
		conf.MsgChanCap = defaultMsgChCap
	}
	if conf.ErrChanCap < 0 {
		conf.ErrChanCap = defaultErrChCap
	}
	if conf.NumMessagesPerSecond < 0 {
		conf.NumMessagesPerSecond = defaultNumMessagesPerSecond
	}
	return conf
}
