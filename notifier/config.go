package notifier

import (
	"fmt"
	"strings"
)

type Config struct {
	NumWorkers  int
	MaxChCap    int
	MaxErrChCap int
}

func (c Config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{\n"))
	sb.WriteString(fmt.Sprintf("  NumWorkers: %v,\n", c.NumWorkers))
	sb.WriteString(fmt.Sprintf("  MaxChCap: %v,\n", c.MaxChCap))
	sb.WriteString(fmt.Sprintf("  MaxErrChCap: %v,\n", c.MaxErrChCap))
	sb.WriteString(fmt.Sprintf("}\n"))
	return sb.String()
}
