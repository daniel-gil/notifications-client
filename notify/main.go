package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var interval time.Duration

func main() {
	url, interval, err := parseFlags()
	if err != nil {
		return
	}
	fmt.Printf("Notification client started: URL=%s, interval=%v\n", url, interval)

	initSignalsHandler()

	for {
		// TODO: read each 'interval' from the stdin

		// TODO: handle messages
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

			fmt.Printf("\nTermination signal caught: %v.\nAre you sure that you want to terminate the application? [y/n]", sig)

			// read the user response from the standard input (keyboard)
			var s string
			_, err := fmt.Scan(&s)
			if err != nil {
				panic(err)
			}
			s = strings.TrimSpace(s)
			s = strings.ToLower(s)

			// in case the response is positive, we should finish the program
			if s == "y" || s == "yes" {
				fmt.Println("Exit program")
				os.Exit(0)
			}
			fmt.Println("Termination cancelled")
		}
	}()
}
