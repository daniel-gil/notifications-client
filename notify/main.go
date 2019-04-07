package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/daniel-gil/notifications-client/notifier"
	log "github.com/sirupsen/logrus"
)

const defaultInterval = 5 * time.Second
const defaultChannelCapacity = 500
const defaultMaxNumRetrials = 2
const defaultMaxNumMessagesToProcess = 1000

var client notifier.Notifier
var conf *config

type config struct {
	url                     string
	interval                time.Duration
	channelCapacity         int
	maxNumRetrials          int
	maxNumMessagesToProcess int
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

func main() {
	// read parameters and arguments from flags
	err := parseFlags()
	if err != nil {
		return
	}
	log.Infof("HTTP Notification client started\nListening for new messages using the following configuration: \n%v\n", conf)

	// create a goroutine dedicated to read lines from stdin and send them to a channel to be processed later (each interval)
	ch := messageReader(os.Stdin)

	client = notifier.New(conf.url)
	go initErrorHandler()

	// read each 'interval' from the stdin and send the notifications using the notifier library
	t := time.Tick(conf.interval)
	for range t {
		processMessages(ch)
	}
}

func init() {
	// logrus configuration
	initLogger()

	// configuration to handle the SIGINT termination signal
	initSignalsHandler()
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
		msg := <-ch
		if len(msg) > 0 {
			messages = append(messages, msg)
		}
	}

	if len(messages) == 0 {
		log.Debugf("no new messages")
	} else {
		// send those messages to the notifier client
		guid, err := client.Notify(messages)
		if err != nil {
			log.Errorf("notifier has reported a failure: %v", err)
		}
		log.Infof("messages notified: GUID=%s", guid)
	}
}

func parseFlags() error {
	conf = &config{}

	const (
		urlFlagUsage                     = "URL where to send notifications"
		intervalFlagUsage                = "Notification interval"
		channelCapacityFlagUsage         = "Channel capacity for reading from stdin"
		maxNumRetrialsFlagUsage          = "Maximal number of retrials when receives an error sending a notification"
		maxNumMessagesToProcessFlagUsage = "Maximal number of messages to be processed per interval"
	)

	// display a usage text if no parameters
	if len(os.Args) == 1 {
		fmt.Printf("usage: notify --url=URL [<flags>]\n")
		fmt.Printf("\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("	--help			Shows context-sensitive help\n")
		fmt.Printf("	-i, --interval=5s	%s\n", intervalFlagUsage)
		fmt.Printf("	-c, --chcap=500		%s\n", channelCapacityFlagUsage)
		fmt.Printf("	-r, --retrials=2	%s\n", maxNumRetrialsFlagUsage)
		fmt.Printf("	-m, --messages=1000	%s\n", maxNumMessagesToProcessFlagUsage)
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

	// parse the flags previously defined
	flag.Parse()

	// check that we received all mandatory parameters
	if conf.url == "" {
		fmt.Printf("missing URL parameter\n")
		return fmt.Errorf("missing URL parameter")
	}
	return nil
}

func initSignalsHandler() {
	//  create a channel to receive OS signal notifications
	sigs := make(chan os.Signal, 1)

	// register the channel `sigs` for receiving the specified notifications
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// start a goroutine to handle the termination signal
	go func() {
		for {
			// remains blocked here until a termination signal is received and read from the channel
			sig := <-sigs

			log.Printf("\nSignal caught: %+v\nExit program\n", sig)
			os.Exit(0)
		}
	}()
}

func initLogger() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func messageReader(r io.Reader) <-chan string { // returns receive-only channel of strings
	ch := make(chan string, defaultChannelCapacity)

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

func initErrorHandler() {
	errCh := client.GetErrorChannel()
	for {
		select {
		case e := <-errCh:
			log.Errorf("Handling new error: [%s] for message: { GUID : \"%s\", Index : %v, Content : \"%s\" }", e.Error, e.GUID, e.Index, e.Message)

			if e.NumRetrials < conf.maxNumRetrials {
				// retry to send this failed notification
				log.Warnf("Retrial[%v]: { GUID : \"%s\", Index : %v, Content : \"%s\" }", e.NumRetrials, e.GUID, e.Index, e.Message)
				client.Retry(e.Message, e.GUID, e.Index, e.NumRetrials)
			}
		}
	}
}
