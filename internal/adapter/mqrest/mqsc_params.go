package mqrest

import (
	"strconv"
	"strings"

	"github.com/konradheimel/kurator/internal/mqadmin"
)

const attrMaxDepth = "maxdepth"

// queueDisplayParameters lists attributes safe for runCommandJSON DISPLAY qlocal
// on IBM MQ 9.4.x. Some keywords (e.g. maxmsglen) are rejected by mqweb with
// MQWB0120E even though they are valid on DEFINE.
var queueDisplayParameters = []string{
	attrMaxDepth, "descr", "defpsist", "get", "put",
}

// queueNumericParameters are coerced to JSON numbers for runCommandJSON DEFINE.
var queueNumericParameters = map[string]struct{}{
	attrMaxDepth: {},
	"maxmsglen":  {},
}

func defineQueueParameters(spec mqadmin.QueueSpec) map[string]any {
	params := map[string]any{"replace": "yes"}
	for k, v := range spec.Attributes {
		key := strings.ToLower(k)
		if _, numeric := queueNumericParameters[key]; numeric {
			if n, err := strconv.Atoi(v); err == nil {
				params[key] = n
				continue
			}
		}
		params[key] = v
	}
	return params
}
