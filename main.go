package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	ntfr "github.com/daniel-gil/notifications-client/notifier"
	log "github.com/sirupsen/logrus"
)

const maxChannelCapacity = 500
const maxNumRetrials = 2

var interval time.Duration
var notifier ntfr.Notifier

func main() {
	// read parameters and arguments from flags
	url, interval, err := parseFlags()
	if err != nil {
		return
	}
	log.Infof("HTTP Notification client started\nListening for new messages at URL %s with interval %v\n", url, interval)

	// create a goroutine dedicated to read lines from stdin and send them to a channel to be processed later (each interval)
	ch := messageReader(os.Stdin)

	notifier = ntfr.New(url)
	go initErrorHandler()

	// read each 'interval' from the stdin and send the notifications using the notifier library
	t := time.Tick(*interval)
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
	messages := []string{}
	for i := 0; i < numMsgs; i++ {
		messages = append(messages, <-ch)
	}

	if len(messages) == 0 {
		log.Debugf("no new messages")
	} else {
		log.Debugf("new messages read: %v", messages)

		// send those messages to the notifier client
		guid, err := notifier.Notify(messages)
		if err != nil {
			log.Errorf("notifier has reported a failure: %v", err)
		}
		log.Infof("messages notified: GUID=%s", guid)
	}
}

func parseFlags() (string, *time.Duration, error) {
	// display a usage text if no parameters
	if len(os.Args) == 1 {
		fmt.Println("usage: notify --url=URL [<flags>]")
		fmt.Println("")
		fmt.Println("Flags:")
		fmt.Println("	--help			Shows context-sensitive help")
		fmt.Println("	-i, --interval=5s	Notification interval")
		return "", nil, fmt.Errorf("wrong usage")
	}

	// define the URL flag
	url := flag.String("url", "", "URL where to send notifications")

	// define the interval flag (admits also the short alternative form)
	const (
		flagValue = 5 * time.Second
		flagUsage = "Notification interval"
	)
	flag.DurationVar(&interval, "interval", flagValue, flagUsage)
	flag.DurationVar(&interval, "i", flagValue, flagUsage+" (shorthand)")

	// parse the flags previously defined
	flag.Parse()

	// check that we received all mandatory parameters
	if url == nil || *url == "" {
		fmt.Printf("missing URL parameter\n")
		return "", nil, fmt.Errorf("missing URL parameter")
	}
	return *url, &interval, nil
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
	ch := make(chan string, maxChannelCapacity)

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
	errCh := notifier.GetErrorChannel()
	for {
		select {
		case e := <-errCh:
			log.Errorf("Handling new error: [%s] for message: { GUID : \"%s\", Index : %v, Content : \"%s\" }", e.Error, e.GUID, e.Index, e.Message)

			if e.NumRetrials < maxNumRetrials {
				// retry to send this failed notification
				log.Warnf("Retrial[%v]: { GUID : \"%s\", Index : %v, Content : \"%s\" }", e.NumRetrials, e.GUID, e.Index, e.Message)
				notifier.Retry(e.Message, e.GUID, e.Index, e.NumRetrials)
			}
		}
	}
}
