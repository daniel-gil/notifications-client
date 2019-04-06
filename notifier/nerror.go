package notifier

type NError struct {
	Error       string
	Message     string
	NumRetrials int
	GUID        string
	Index       int
}
