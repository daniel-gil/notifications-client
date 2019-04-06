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

	"github.com/daniel-gil/notifications-client/client"
	log "github.com/sirupsen/logrus"
)

var interval time.Duration

const maxStdinChannelCapacity = 1000

var stopping = false

func main() {
	initLogger()

	// read parameters and arguments from flags
	url, interval, err := parseFlags()
	if err != nil {
		return
	}
	log.Infof("Notification client started: URL=%s, interval=%v\n", url, interval)

	// configuration to handle the SIGINT termination signal
	initSignalsHandler()

	notifier := client.New()

	// create a goroutine dedicated to read lines from stdin and send them to a channel to be processed later (each interval)
	ch := make(chan string, maxStdinChannelCapacity)
	go messageReader(os.Stdin, ch)

	// read each 'interval' from the stdin
	t := time.Tick(*interval)
	for range t {
		if !stopping {
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

				// TODO: send those messages to the notifier client
				guid, err := notifier.Notify(messages)
				if err != nil {
					log.Fatalf("notifier has reported a failure: %v", err)
				}
				log.Infof("messages notified: GUID=%s", guid)
			}
		}
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

func messageReader(r io.Reader, stdinCh chan string) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if !stopping {
			line := scanner.Text()
			stdinCh <- line
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed at scanning stdin: %s", err)
	}
}
