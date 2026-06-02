package controller

import (
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
	"github.com/konih/kurator/internal/mqadmin"
)

const (
	EventReasonConnectionNotFound = "ConnectionNotFound"
	EventReasonSecretNotFound     = "SecretNotFound"
	EventReasonDeleted            = "Deleted"
)

func classifyReconcileError(err error) (reason, message string) {
	message = err.Error()

	var term *mqadmin.TerminalError
	if errors.As(err, &term) {
		reason = messagingv1alpha1.ReasonError
		if term.Reason != "" {
			reason = term.Reason
		}
		if term.Message != "" {
			message = term.Message
		}
		return reason, message
	}

	if hasNotFoundInChain(err) {
		if containsInChain(err, "get connection") {
			return EventReasonConnectionNotFound, message
		}
		if containsInChain(err, "credentials secret") || containsInChain(err, "CA secret") {
			return EventReasonSecretNotFound, message
		}
	}

	return messagingv1alpha1.ReasonError, message
}

func recordReconcileWarning(recorder record.EventRecorder, obj runtime.Object, err error) {
	if recorder == nil || err == nil || errors.Is(err, mqadmin.ErrTransient) {
		return
	}
	reason, message := classifyReconcileError(err)
	recorder.Eventf(obj, corev1.EventTypeWarning, reason, "%s", message)
}

func recordNormalEvent(recorder record.EventRecorder, obj runtime.Object, reason, message string) {
	if recorder == nil {
		return
	}
	recorder.Eventf(obj, corev1.EventTypeNormal, reason, "%s", message)
}

func conditionChanged(
	conditions []metav1.Condition,
	condType string,
	newStatus metav1.ConditionStatus,
	newReason string,
) bool {
	for _, c := range conditions {
		if c.Type == condType {
			return c.Status != newStatus || c.Reason != newReason
		}
	}
	return true
}

func hasNotFoundInChain(err error) bool {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if k8serrors.IsNotFound(e) {
			return true
		}
	}
	return false
}

func containsInChain(err error, substr string) bool {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if strings.Contains(e.Error(), substr) {
			return true
		}
	}
	return false
}
