package notifier

type message struct {
	content     string
	guid        string
	numRetrials int
	index       int
}
