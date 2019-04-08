package notilib

import (
	"fmt"
	"strings"
)

type Config struct {
	BurstLimit           int
	NumMessagesPerSecond int
	MsgChanCap           int
	MaxErrChCap          int
}

func (c Config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{\n"))
	sb.WriteString(fmt.Sprintf("  BurstLimit: %v,\n", c.BurstLimit))
	sb.WriteString(fmt.Sprintf("  NumMessagesPerSecond: %v,\n", c.NumMessagesPerSecond))
	sb.WriteString(fmt.Sprintf("  MsgChanCap: %v,\n", c.MsgChanCap))
	sb.WriteString(fmt.Sprintf("  MaxErrChCap: %v,\n", c.MaxErrChCap))
	sb.WriteString(fmt.Sprintf("}\n"))
	return sb.String()
}
