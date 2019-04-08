package notilib

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type sender interface {
	send(msg message)
}

type senderHandler struct {
	url    string
	client dispatcher
	errCh  chan NError
}

func newSender(url string, client dispatcher, errCh chan NError) sender {
	return &senderHandler{
		url:    url,
		client: client,
		errCh:  errCh,
	}
}

// send is responsible for sending the request to the client
func (f *senderHandler) send(msg message) {
	body := strings.NewReader(msg.content)
	req, err := http.NewRequest("POST", f.url, body)
	resp, err := f.client.dispatch(req)
	if err != nil {
		f.reportError(msg, fmt.Errorf("unable to send the request: %v", err))
		return
	}

	// defer the close operation of the response body to avoid a resource leak
	defer resp.Body.Close()

	// check if the response is a successful HTTP code: 200 OK or 201 Created
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		f.reportError(msg, fmt.Errorf("unexpected HTTP Status: %s", resp.Status))
		return
	}
	log.Debugf("Message sent correctly: HttpCode=%s, GUID=[%s], index=%d", resp.Status, msg.guid, msg.index)
}

func (f *senderHandler) reportError(msg message, err error) {
	f.errCh <- NError{
		GUID:        msg.guid,
		Index:       msg.index,
		Error:       err.Error(),
		Message:     msg.content,
		NumRetrials: msg.numRetrials,
	}
}
