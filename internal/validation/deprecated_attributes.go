package validation

import (
	"fmt"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

const deprecatedAttrWarningFmt = `spec.attributes[%q] is deprecated; use %s`

var (
	deprecatedQueueAttributeFields = map[string]string{
		"maxdepth":          "spec.maxDepth",
		"descr":             "spec.description",
		"defpsist":          "spec.defPersistence",
		"get":               "spec.get",
		"put":               "spec.put",
		"targq":             "spec.targetQueue",
		"target":            "spec.targetQueue",
		"xmitq":             "spec.xmitQueue",
		"transmissionqueue": "spec.xmitQueue",
		"rqmname":           "spec.remoteQueueManager",
		"remotemanager":     "spec.remoteQueueManager",
	}
	deprecatedTopicAttributeFields = map[string]string{
		"topstr":   "spec.topicString",
		"topicstr": "spec.topicString",
		"descr":    "spec.description",
		"pub":      "spec.publish",
		"sub":      "spec.subscribe",
		"defpsist": "spec.defPersistence",
		"pubscope": "spec.publishScope",
		"subscope": "spec.subscribeScope",
	}
	deprecatedChannelAttributeFields = map[string]string{
		"descr":    "spec.description",
		"maxmsgl":  "spec.maxMsgLength",
		"trptype":  "spec.transportType",
		"sharecnv": "spec.shareConv",
		"mcauser":  "spec.mcaUser",
		"maxinst":  "spec.maxInstances",
		"maxinstc": "spec.maxInstancesClient",
		"sslciph":  "spec.sslCipherSpec",
		"sslcauth": "spec.sslClientAuth",
		"conname":  "spec.connName",
		"xmitq":    "spec.xmitQueue",
	}
)

func deprecatedQueueAttributeWarnings(attrs map[string]string) []string {
	return deprecatedAttributeWarnings(attrs, deprecatedQueueAttributeFields)
}

func deprecatedTopicAttributeWarnings(attrs map[string]string) []string {
	return deprecatedAttributeWarnings(attrs, deprecatedTopicAttributeFields)
}

func deprecatedChannelAttributeWarnings(attrs map[string]string) []string {
	return deprecatedAttributeWarnings(attrs, deprecatedChannelAttributeFields)
}

func deprecatedAttributeWarnings(attrs map[string]string, replacements map[string]string) []string {
	if len(attrs) == 0 {
		return nil
	}
	warnings := make([]string, 0, len(attrs))
	for key := range attrs {
		replacement, ok := replacements[mqadmin.NormalizeAttrKey(key)]
		if !ok {
			continue
		}
		warnings = append(warnings, fmt.Sprintf(deprecatedAttrWarningFmt, key, replacement))
	}
	return warnings
}
