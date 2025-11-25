package eap

import (
	"fmt"
	"strconv"
	"strings"
)

// String returns a human readable representation of the EAP message.
func (msg *EAPMessage) String() string {
	if msg == nil {
		return "EAPMessage<nil>"
	}

	var builder strings.Builder
	builder.WriteString("EAPMessage{")
	builder.WriteString("Code=")
	builder.WriteString(strconv.FormatUint(uint64(msg.Code), 10))
	builder.WriteString(", Identifier=")
	builder.WriteString(strconv.FormatUint(uint64(msg.Identifier), 10))
	builder.WriteString(", Length=")
	builder.WriteString(strconv.FormatUint(uint64(msg.Length), 10))
	builder.WriteString(", Type=")
	builder.WriteString(strconv.FormatUint(uint64(msg.Type), 10))
	builder.WriteString(", Data=")
	builder.WriteString(fmt.Sprintf("%x", msg.Data))
	builder.WriteString("}")
	return builder.String()
}
