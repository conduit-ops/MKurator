package validation

import (
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation/field"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
)

var mqAuthorityKeywordPattern = regexp.MustCompile(`^[A-Za-z0-9+_]+$`)

var validChannelAuthUserSources = map[messagingv1alpha1.ChannelAuthUserSource]struct{}{
	messagingv1alpha1.ChannelAuthUserSourceChannel:  {},
	messagingv1alpha1.ChannelAuthUserSourceNoAccess: {},
}

var validChannelAuthCheckClients = map[messagingv1alpha1.ChannelAuthCheckClient]struct{}{
	messagingv1alpha1.ChannelAuthCheckClientRequired: {},
	messagingv1alpha1.ChannelAuthCheckClientAsQMGR:   {},
	messagingv1alpha1.ChannelAuthCheckClientReqdAdm:  {},
	messagingv1alpha1.ChannelAuthCheckClientAsCHL:    {},
	messagingv1alpha1.ChannelAuthCheckClientOptional: {},
}

// ValidateChannelAuthUserSource checks USERSRC for ADDRESSMAP rules when set.
func ValidateChannelAuthUserSource(path *field.Path, value messagingv1alpha1.ChannelAuthUserSource) field.ErrorList {
	if value == "" {
		return nil
	}
	if _, ok := validChannelAuthUserSources[value]; ok {
		return nil
	}
	return field.ErrorList{
		field.Invalid(path, value, "must be one of: CHANNEL, NOACCESS"),
	}
}

// ValidateChannelAuthCheckClient checks CHCKCLNT for ADDRESSMAP rules when set.
func ValidateChannelAuthCheckClient(path *field.Path, value messagingv1alpha1.ChannelAuthCheckClient) field.ErrorList {
	if value == "" {
		return nil
	}
	if _, ok := validChannelAuthCheckClients[value]; ok {
		return nil
	}
	return field.ErrorList{
		field.Invalid(path, value, "must be one of: REQUIRED, ASQMGR, REQDADM, ASCHL, OPTIONAL"),
	}
}

// ValidateMQAuthorityKeyword checks a single AUTHADD authority keyword.
func ValidateMQAuthorityKeyword(path *field.Path, value string) field.ErrorList {
	if mqAuthorityKeywordPattern.MatchString(value) {
		return nil
	}
	return field.ErrorList{
		field.Invalid(path, value, "authority must match ^[A-Za-z0-9+_]+$"),
	}
}
