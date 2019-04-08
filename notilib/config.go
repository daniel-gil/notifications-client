package notilib

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Config is the configuration for initializing the notilib. It is optional, if nil is passed, default values will be used.
type Config struct {
	BurstLimit           int       // Burst limit for the listener, allowing to process several messages from the Message Channel per rate
	NumMessagesPerSecond int       // Maximal number of messages to be processed per second (it will be used to calculate the rate for the rate limiter)
	MsgChanCap           int       // Message Channel Capacity
	ErrChanCap           int       // Error Channel Capacity
	LogLevel             log.Level // log level for logrus
}

func (c Config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{\n"))
	sb.WriteString(fmt.Sprintf("  BurstLimit: %v,\n", c.BurstLimit))
	sb.WriteString(fmt.Sprintf("  NumMessagesPerSecond: %v,\n", c.NumMessagesPerSecond))
	sb.WriteString(fmt.Sprintf("  MsgChanCap: %v,\n", c.MsgChanCap))
	sb.WriteString(fmt.Sprintf("  ErrChanCap: %v,\n", c.ErrChanCap))
	sb.WriteString(fmt.Sprintf("  LogLevel: %v,\n", c.LogLevel))
	sb.WriteString(fmt.Sprintf("}\n"))
	return sb.String()
}
