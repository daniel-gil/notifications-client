package client

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

type Notifier interface {
	Notify(messages []string) (string, error)
}

type notifier struct {
}

func New() Notifier {
	notifier := &notifier{}
	return notifier
}

func (n *notifier) Notify(messages []string) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("unable to create an GUID: %v", err)
	}
	return id.String(), nil
}
