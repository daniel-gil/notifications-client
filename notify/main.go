package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var interval time.Duration

const maxStdinChannelCapacity = 1000

var stopping = false

func main() {
	// read parameters and arguments from flags
	url, interval, err := parseFlags()
	if err != nil {
		return
	}
	fmt.Printf("Notification client started: URL=%s, interval=%v\n", url, interval)

	// configuration to handle the SIGINT termination signal
	initSignalsHandler()

	// create a goroutine dedicated to read lines from stdin and send them to a channel to be processed later (each interval)
	stdinCh := make(chan string, maxStdinChannelCapacity)
	go stdinReader(stdinCh)

	// read each 'interval' from the stdin
	c := time.Tick(*interval)
	for now := range c {
		if !stopping {
			numMsgs := len(stdinCh)
			fmt.Printf("\n%vNew tick, num messages=%v", now, numMsgs)
			messages := []string{}
			for i := 0; i < numMsgs; i++ {
				messages = append(messages, <-stdinCh)
			}
			fmt.Printf("\n%v New messages read: %v\n", now, messages)

			// TODO: send those messages to the notifier client
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

			fmt.Printf("\nSignal caught: %+v\nExit program\n", sig)
			os.Exit(0)
		}
	}()
}

func stdinReader(stdinCh chan string) {
	scanner := bufio.NewScanner(os.Stdin)
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
