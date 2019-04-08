package notilib

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

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

const defaultMaxChCap = 1000
const defaultMaxErrChCap = 500
const defaultBurstLimit = 1000
const defaultNumMessagesPerSecond = 1000
const defaultReqChanCapacity = 100

// New creates a new object that implements Notilib interface
func New(url string, client *http.Client, conf *Config) (Notilib, error) {
	// validate the URL format
	err := checkURLFormat(url)
	if err != nil {
		return nil, err
	}

	// if no configuration is provided, build a default configuration
	conf = getConfiguration(conf)
	log.Debugf("Notifier configuration: \n%v\n", conf)

	// create channels
	msgChan := make(chan message, conf.MaxChCap)
	errCh := make(chan NError, conf.MaxErrChCap)

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
