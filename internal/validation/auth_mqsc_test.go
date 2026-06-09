package validation

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
)

func authValidationClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()
}

func TestValidateChannelAuthRuleSpecRejectsMQSCInjectionUserSource(t *testing.T) {
	t.Parallel()
	cl := channelAuthRuleValidationClient(t)

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "inject-car",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
			UserSource:    `MAP) MCAUSER('mqm'`,
		})
	if !channelAuthRuleFieldError(errs, field.ErrorTypeInvalid, "spec.userSource") {
		t.Fatalf("expected spec.userSource invalid, got %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecRejectsMQSCInjectionCheckClient(t *testing.T) {
	t.Parallel()
	cl := channelAuthRuleValidationClient(t)

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "inject-car",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
			CheckClient:   "REQUIRED) ACTION(REPLACE",
		})
	if !channelAuthRuleFieldError(errs, field.ErrorTypeInvalid, "spec.checkClient") {
		t.Fatalf("expected spec.checkClient invalid, got %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecUserSourceCheckClientEnumTable(t *testing.T) {
	t.Parallel()
	cl := channelAuthRuleValidationClient(t)

	base := func() *messagingv1alpha1.ChannelAuthRuleSpec {
		return &messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		}
	}

	tests := []struct {
		name      string
		mutate    func(*messagingv1alpha1.ChannelAuthRuleSpec)
		wantField string
	}{
		{
			name: "userSource CHANNEL allowed",
			mutate: func(s *messagingv1alpha1.ChannelAuthRuleSpec) {
				s.UserSource = messagingv1alpha1.ChannelAuthUserSourceChannel
			},
		},
		{
			name: "userSource NOACCESS allowed",
			mutate: func(s *messagingv1alpha1.ChannelAuthRuleSpec) {
				s.UserSource = messagingv1alpha1.ChannelAuthUserSourceNoAccess
			},
		},
		{
			name: "checkClient REQUIRED allowed",
			mutate: func(s *messagingv1alpha1.ChannelAuthRuleSpec) {
				s.CheckClient = messagingv1alpha1.ChannelAuthCheckClientRequired
			},
		},
		{
			name: "checkClient ASQMGR allowed",
			mutate: func(s *messagingv1alpha1.ChannelAuthRuleSpec) {
				s.CheckClient = messagingv1alpha1.ChannelAuthCheckClientAsQMGR
			},
		},
		{
			name: "userSource lowercase rejected",
			mutate: func(s *messagingv1alpha1.ChannelAuthRuleSpec) {
				s.UserSource = messagingv1alpha1.ChannelAuthUserSource("channel")
			},
			wantField: "spec.userSource",
		},
		{
			name: "checkClient garbage rejected",
			mutate: func(s *messagingv1alpha1.ChannelAuthRuleSpec) {
				s.CheckClient = messagingv1alpha1.ChannelAuthCheckClient("NOTREAL")
			},
			wantField: "spec.checkClient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spec := base()
			tt.mutate(spec)
			errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car-enum", spec)
			if tt.wantField == "" {
				if len(errs) > 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if !channelAuthRuleFieldError(errs, field.ErrorTypeInvalid, tt.wantField) {
				t.Fatalf("expected %s invalid, got %v", tt.wantField, errs)
			}
		})
	}
}

func TestValidateAuthorityRecordSpecRejectsMQSCInjectionAuthority(t *testing.T) {
	t.Parallel()
	cl := authValidationClient(t)

	errs := ValidateAuthorityRecordSpec(context.Background(), cl, "default", "inject-auth",
		&messagingv1alpha1.AuthorityRecordSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{`GET) AUTHADD(ALL`},
		})
	if !authorityRecordFieldError(errs, field.ErrorTypeInvalid, "spec.authorities[0]") {
		t.Fatalf("expected spec.authorities[0] invalid, got %v", errs)
	}
}

func TestValidateAuthorityRecordSpecAuthorityPatternTable(t *testing.T) {
	t.Parallel()
	cl := authValidationClient(t)

	base := func(authorities ...string) *messagingv1alpha1.AuthorityRecordSpec {
		return &messagingv1alpha1.AuthorityRecordSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   authorities,
		}
	}

	tests := []struct {
		name      string
		spec      *messagingv1alpha1.AuthorityRecordSpec
		wantField string
	}{
		{
			name: "GET PUT CONNECT allowed",
			spec: base("GET", "PUT", "CONNECT"),
		},
		{
			name: "SETALL allowed",
			spec: base("SETALL"),
		},
		{
			name:      "authority with space rejected",
			spec:      base("GET PUT"),
			wantField: "spec.authorities[0]",
		},
		{
			name:      "authority with paren rejected",
			spec:      base("GET)"),
			wantField: "spec.authorities[0]",
		},
		{
			name:      "authority with quote rejected",
			spec:      base("GET'"),
			wantField: "spec.authorities[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateAuthorityRecordSpec(context.Background(), cl, "default", "auth-pattern", tt.spec)
			if tt.wantField == "" {
				if len(errs) > 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if !authorityRecordFieldError(errs, field.ErrorTypeInvalid, tt.wantField) {
				t.Fatalf("expected %s invalid, got %v", tt.wantField, errs)
			}
		})
	}
}

func authorityRecordFieldError(errs field.ErrorList, typ field.ErrorType, fld string) bool {
	for _, err := range errs {
		if err.Type == typ && err.Field == fld {
			return true
		}
	}
	return false
}
