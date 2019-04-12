package notilib

import "fmt"

// NError struct sent to the Error Channel
type NError struct {
	ErrorMessage string // Error message
	Content      string // Original failed notification
	NumRetrials  int    // Number of retrials
	GUID         string // GUID: Unique identifier
	Index        int    // Index of the message from the []string passed as parameter to the notilib.Notify method
}

// implementing the error interface
func (e NError) Error() string {
	return fmt.Sprintf("[%s][%d]: \"%s\" for message: \"%s\" ", e.GUID, e.Index, e.ErrorMessage, e.Content)
}
