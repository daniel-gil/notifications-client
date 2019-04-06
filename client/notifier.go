package client

import (
	"fmt"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type Notifier interface {
	Notify(messages []string) (string, error)
}

type notifier struct {
	url string
}

func New(url string) Notifier {
	notifier := &notifier{
		url: url,
	}
	return notifier
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

func (n *notifier) Notify(messages []string) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("unable to create an GUID: %v", err)
	}

	for i, msg := range messages {
		body := strings.NewReader(msg)
		req, err := http.NewRequest("POST", n.url, body)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("unable to send the request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Warnf("GUID=[%s] index=%v, HTTP Status=%s", id, i, resp.Status)
		}
	}
	return id.String(), nil
}
