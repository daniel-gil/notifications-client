package notilib

// NError struct sent to the Error Channel
type NError struct {
	Error       string // Error message
	Message     string // Original failed notification
	NumRetrials int    // Number of retrials
	GUID        string // GUID: Unique identifier
	Index       int    // Index of the message from the []string passed as parameter to the notilib.Notify method
}
