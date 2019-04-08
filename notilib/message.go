package notilib

type message struct {
	content     string // notification text message
	guid        string // GUID: Unique identifier
	index       int    // Index of the message from the []string passed as parameter to the notilib.Notify method
	numRetrials int    // Current number of retrials for this notification
}
