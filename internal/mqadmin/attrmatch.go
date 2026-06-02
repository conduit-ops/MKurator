package mqadmin

import (
	"strconv"
	"strings"
)

var caseInsensitiveAttrKeys = map[string]struct{}{
	"pub":      {},
	"sub":      {},
	"get":      {},
	"put":      {},
	"defpsist": {},
	"trptype":  {},
}

var numericAttrKeys = map[string]struct{}{
	"maxdepth":  {},
	"maxmsglen": {},
	"maxmsgl":   {},
	"sharecnv":  {},
	"maxinst":   {},
	"maxinstc":  {},
}

// AttributeValueMatches reports whether desired and observed MQ attribute values
// are equivalent for drift detection.
func AttributeValueMatches(key, desired, observed string) bool {
	key = strings.ToLower(key)
	if key == "topicstr" {
		key = "topstr"
	}
	if _, ok := caseInsensitiveAttrKeys[key]; ok {
		return strings.EqualFold(strings.TrimSpace(desired), strings.TrimSpace(observed))
	}
	if _, ok := numericAttrKeys[key]; ok {
		return normalizeNumericString(desired) == normalizeNumericString(observed)
	}
	return strings.TrimSpace(desired) == strings.TrimSpace(observed)
}

// AttributesNeedUpdate returns true when any desired attribute differs from observed.
func AttributesNeedUpdate(desired map[string]string, observed map[string]string) bool {
	for k, v := range desired {
		key := strings.ToLower(k)
		if key == "topicstr" {
			key = "topstr"
		}
		if !AttributeValueMatches(key, v, observed[key]) {
			return true
		}
	}
	return false
}

func normalizeNumericString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	return strconv.Itoa(n)
}
