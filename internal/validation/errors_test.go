package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestChannelAuthRuleInvalid(t *testing.T) {
	err := ChannelAuthRuleInvalid("car1", field.ErrorList{
		field.Required(field.NewPath("spec").Child("address"), "required"),
	})
	if err == nil {
		t.Fatal("expected invalid error")
	}
}

func TestAuthorityRecordInvalid(t *testing.T) {
	err := AuthorityRecordInvalid("auth1", field.ErrorList{
		field.Required(field.NewPath("spec").Child("principal"), "required"),
	})
	if err == nil {
		t.Fatal("expected invalid error")
	}
}
