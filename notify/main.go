package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	nl "github.com/daniel-gil/notifications-client/notilib"
	log "github.com/sirupsen/logrus"
)

const defaultInterval = 5 * time.Second
const defaultChannelCapacity = 500
const defaultMaxNumRetrials = 2
const defaultMaxNumMessagesToProcess = 100
const defaultLogLevel = log.InfoLevel

var notilib nl.Notilib
var conf *config

type config struct {
	url                     string
	interval                time.Duration
	channelCapacity         int
	maxNumRetrials          int
	maxNumMessagesToProcess int
	logLevel                log.Level
}

func main() {
	// read parameters and arguments from flags
	err := parseFlags()
	if err != nil {
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// init logger and configure signals notifications (SIGINT)
	setup(conf.logLevel, cancel)

	log.Infof("HTTP Notification client started. Listening for new messages from stdin...")
	log.Debugf("Notify configuration: \n%v\n", conf)

	// create a goroutine dedicated to read lines from stdin and send them to a channel to be processed later (each interval)
	ch := listen(os.Stdin, conf.channelCapacity)

	// create a notilib instance using the default configuration (passing nil as the second parameter)
	config := nl.DefaultConfig()
	config.LogLevel = conf.logLevel
	notilib, err = nl.New(conf.url, http.DefaultClient, config)
	if err != nil {
		log.Errorf("unable to start the client: %v", err)
		return
	}

	// start the error handler responsible for retrials
	initErrorHandler()

	notilib.Listen(ctx)

	// process messages each 'interval'
	tick := time.Tick(conf.interval)
	for {
		<-tick
		processMessages(ch)
	}
}

func setup(logLevel log.Level, cancel context.CancelFunc) {
	// logrus configuration
	initLogger(logLevel)

	// configuration to handle the SIGINT termination signal
	initSignalsHandler(cancel)
}

// listen to the stdin capturing all the messages and inserting them into the Stdin Channel
func listen(r io.Reader, chanCap int) <-chan string {
	ch := make(chan string, chanCap)

	go func() {
		scanner := bufio.NewScanner(r)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := scanner.Text()
			ch <- line
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("failed at scanning stdin: %s", err)
		}
	}()

	return ch
}

func processMessages(ch <-chan string) {
	numMsgs := len(ch)
	log.Debugf("new tick. Num messages in channel: %v", numMsgs)

	// control the maximal amount of messages to be procesed each interval
	if numMsgs > conf.maxNumMessagesToProcess {
		numMsgs = conf.maxNumMessagesToProcess
	}

	messages := []string{}
	for i := 0; i < numMsgs; i++ {
		msg, ok := <-ch
		if !ok {
			log.Fatal("Stdin Channel is closed unexpectedly")
		}
		if len(msg) > 0 {
			messages = append(messages, msg)
		}
	}

	if len(messages) == 0 {
		log.Debugf("no new messages")
	} else {
		// send those messages to the notifier client
		guid, err := notilib.Notify(messages)
		if err != nil {
			log.Errorf("notifier client has reported a failure: %v", err)
		}
		log.Infof("messages received: GUID=%s", guid)
	}
}

func parseFlags() error {
	conf = &config{}
	const (
		urlFlagUsage                     = "URL where to send notifications"
		intervalFlagUsage                = "Notification interval"
		channelCapacityFlagUsage         = "Stdin Channel capacity for reading messages from stdin"
		maxNumRetrialsFlagUsage          = "Maximal number of retrials when receives an error sending a notification"
		maxNumMessagesToProcessFlagUsage = "Maximal number of messages to be processed per interval"
		logLevelFlagUsage                = "Log level. Valid values: trace, debug, info, warn, error, panic, fatal"
	)

	// display a usage text if no parameters
	if len(os.Args) == 1 {
		fmt.Printf("usage: notify --url=URL [<flags>]\n")
		fmt.Printf("\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("	--help			Shows context-sensitive help\n")
		fmt.Printf("	-i, --interval=%v	%s\n", defaultInterval, intervalFlagUsage)
		fmt.Printf("	-c, --chcap=%d		%s\n", defaultChannelCapacity, channelCapacityFlagUsage)
		fmt.Printf("	-r, --retrials=%d	%s\n", defaultMaxNumRetrials, maxNumRetrialsFlagUsage)
		fmt.Printf("	-m, --messages=%d	%s\n", defaultMaxNumMessagesToProcess, maxNumMessagesToProcessFlagUsage)
		fmt.Printf("	-l, --loglevel=%s	%s\n", defaultLogLevel, logLevelFlagUsage)
		return fmt.Errorf("wrong usage")
	}

	// define the url flag (admits also the short alternative form)
	flag.StringVar(&conf.url, "url", "", urlFlagUsage)
	flag.StringVar(&conf.url, "u", "", urlFlagUsage+" (shorthand)")

	// define the interval flag (admits also the short alternative form)
	flag.DurationVar(&conf.interval, "interval", defaultInterval, intervalFlagUsage)
	flag.DurationVar(&conf.interval, "i", defaultInterval, intervalFlagUsage+" (shorthand)")

	// define the channel capacity flag (admits also the short alternative form)
	flag.IntVar(&conf.channelCapacity, "chcap", defaultChannelCapacity, channelCapacityFlagUsage)
	flag.IntVar(&conf.channelCapacity, "c", defaultChannelCapacity, channelCapacityFlagUsage+" (shorthand)")

	// define the max number of retrials flag (admits also the short alternative form)
	flag.IntVar(&conf.maxNumRetrials, "retrials", defaultMaxNumRetrials, maxNumRetrialsFlagUsage)
	flag.IntVar(&conf.maxNumRetrials, "r", defaultMaxNumRetrials, maxNumRetrialsFlagUsage+" (shorthand)")

	// define the max number of messages to process flag (admits also the short alternative form)
	flag.IntVar(&conf.maxNumMessagesToProcess, "messages", defaultMaxNumMessagesToProcess, maxNumMessagesToProcessFlagUsage)
	flag.IntVar(&conf.maxNumMessagesToProcess, "m", defaultMaxNumMessagesToProcess, maxNumMessagesToProcessFlagUsage+" (shorthand)")

	// define the log level flag (admits also the short alternative form)
	logLevelStr := ""
	flag.StringVar(&logLevelStr, "loglevel", "", logLevelFlagUsage)
	flag.StringVar(&logLevelStr, "l", "", logLevelFlagUsage+" (shorthand)")

	// parse the flags previously defined
	flag.Parse()

	switch logLevelStr {
	case "trace":
		conf.logLevel = log.TraceLevel
	case "debug":
		conf.logLevel = log.DebugLevel
	case "info":
		conf.logLevel = log.InfoLevel
	case "warn":
		conf.logLevel = log.WarnLevel
	case "error":
		conf.logLevel = log.ErrorLevel
	case "panic":
		conf.logLevel = log.PanicLevel
	case "fatal":
		conf.logLevel = log.FatalLevel
	default:
		conf.logLevel = defaultLogLevel
	}

	// check that we received all mandatory parameters
	if conf.url == "" {
		fmt.Printf("missing URL parameter\n")
		return fmt.Errorf("missing URL parameter")
	}
	return nil
}

func initSignalsHandler(cancel context.CancelFunc) {
	//  create a channel to receive OS signal notifications
	sigs := make(chan os.Signal, 1)

	// register the channel `sigs` for receiving the specified notifications
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// start a goroutine to handle the termination signal
	go func(cancel context.CancelFunc) {
		for {
			// remains blocked here until a termination signal is received and read from the channel
			sig := <-sigs
			cancel()
			log.Printf("Signal caught: %+v\nExit program\n", sig)
			os.Exit(0)
		}
	}(cancel)
}

func initLogger(logLevel log.Level) {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(logLevel)
}

func initErrorHandler() {
	errCh := notilib.GetErrorChannel()

	go func(errCh <-chan nl.NError) {
		for {
			select {
			case e, ok := <-errCh:
				if !ok {
					log.Fatalf("Error Channel is closed unexpectedly")
				}
				log.Errorf("Handling new error: [%s] for message: { GUID : \"%s\", Index : %v, Content : \"%s\" }", e.Error, e.GUID, e.Index, e.Message)

				if e.NumRetrials < conf.maxNumRetrials {
					// retry to send this failed notification
					notilib.Retry(e.Message, e.GUID, e.Index, e.NumRetrials)
				}
			}
		}
	}(errCh)
}

func (c config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{\n"))
	sb.WriteString(fmt.Sprintf("  url: \"%s\",\n", c.url))
	sb.WriteString(fmt.Sprintf("  interval: \"%v\",\n", c.interval))
	sb.WriteString(fmt.Sprintf("  channelCapacity: %v,\n", c.channelCapacity))
	sb.WriteString(fmt.Sprintf("  maxNumRetrials: %v,\n", c.maxNumRetrials))
	sb.WriteString(fmt.Sprintf("  maxNumMessagesToProcess: %v,\n", c.maxNumMessagesToProcess))
	sb.WriteString(fmt.Sprintf("}\n"))
	return sb.String()
}
