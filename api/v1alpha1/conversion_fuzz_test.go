package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Native Go fuzz targets for the v1alpha1 (spoke) <-> v1beta1 (hub) conversion of
// all six kinds. Each target builds a spoke object from primitive fuzz arguments
// (Go native fuzzing only accepts scalar args, so a struct cannot be fuzzed
// directly), runs ConvertTo(hub) then ConvertFrom(hub), and asserts:
//
//   - no panic for ANY input (the fuzz engine fails on any panic; no assert needed);
//   - the round-tripped object is reflect.DeepEqual to the original (assertLossless).
//
// All six kinds are lossless BY CONSTRUCTION. The only intentionally-lossy path in
// the converters is the MQSC attribute fold (FoldQueue/Topic/ChannelAttributesToTyped):
// a foldable key such as "maxdepth" in Spec.Attributes is promoted into a typed field
// on the hub and removed from the map, so it does not survive as an attribute on the
// way back. We deliberately DO NOT feed foldable keys: the fuzzed Attributes map uses
// a single guaranteed-non-foldable key ("custom"), mirroring HA-S3's
// TestQueueLosslessRoundTrip. This keeps the strict DeepEqual assertion uniform across
// every kind while still exercising the map-copy path.
//
// The builders emit nil (never empty-but-non-nil) for absent optional maps/slices/
// pointers: CloneStringMap / copyStringSlice / copyConditionsToHub all normalize
// empty->nil, so an empty-non-nil input would fail DeepEqual as a phantom "loss".

// fuzzMeta builds an ObjectMeta from primitive fuzz inputs. Optional maps/slices are
// left nil unless populate is set, matching how copyObjectMeta (DeepCopyInto) treats them.
func fuzzMeta(name, namespace, uid, resourceVersion string, generation int64, populate bool) metav1.ObjectMeta {
	m := metav1.ObjectMeta{
		Name:            name,
		Namespace:       namespace,
		UID:             types.UID(uid),
		ResourceVersion: resourceVersion,
		Generation:      generation,
	}
	if populate {
		m.CreationTimestamp = metav1.Time{Time: time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)}
		m.Labels = map[string]string{"app": name}
		m.Annotations = map[string]string{"note": namespace}
		m.Finalizers = []string{"messaging.mkurator.dev/finalizer"}
	}
	return m
}

// fuzzConditions returns a synced-condition slice when populate is set, else nil.
func fuzzConditions(populate bool, reason, message string) []metav1.Condition {
	if !populate {
		return nil
	}
	return []metav1.Condition{
		{
			Type:               ConditionSynced,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: time.Date(2026, 6, 18, 13, 0, 0, 0, time.UTC)},
			Reason:             reason,
			Message:            message,
		},
	}
}

// fuzzMQStatus returns fully-populated MQObjectStatusFields when populate is set,
// else a zero value (Message "", nil pointers) that survives the round-trip unchanged.
func fuzzMQStatus(populate bool, message string) MQObjectStatusFields {
	if !populate {
		return MQObjectStatusFields{}
	}
	synced := metav1.Time{Time: time.Date(2026, 6, 18, 14, 0, 0, 0, time.UTC)}
	exists := true
	return MQObjectStatusFields{
		Message:        message,
		LastSyncTime:   &synced,
		MQObjectExists: &exists,
	}
}

// fuzzWorkloadPolicies always sets both policy fields; they are plain strings that
// round-trip unchanged (empty strings included).
func fuzzWorkloadPolicies(deletion, adoption string) WorkloadLifecyclePolicies {
	return WorkloadLifecyclePolicies{
		DeletionPolicy: DeletionPolicy(deletion),
		AdoptionPolicy: AdoptionPolicy(adoption),
	}
}

// fuzzAttrs returns a single-entry, non-foldable attribute map when populate is set,
// else nil. The key "custom" is guaranteed not to be folded into any typed field, so
// the map survives the hub round-trip intact.
func fuzzAttrs(populate bool, value string) map[string]string {
	if !populate {
		return nil
	}
	return map[string]string{"custom": value}
}

func fuzzInt32Ptr(set bool, v int32) *int32 {
	if !set {
		return nil
	}
	return &v
}

// FuzzQueueConversionRoundTrip fuzzes Queue ConvertTo/ConvertFrom for panics and loss.
func FuzzQueueConversionRoundTrip(f *testing.F) {
	// Minimal seed: only required fields set, every optional pointer/slice/map nil.
	f.Add("q-min", "default", "", "", int64(0), false,
		"qm1", "APP.MIN", "", "", int32(0), false, "", "", "", "", "", "", "",
		false, "", "", "", "", "", "")
	// Maximal seed: every optional pointer/slice/condition populated.
	f.Add("q-max", "prod", "uid-1", "42", int64(3), true,
		"qm1", "APP.MAX", string(QueueTypeRemote), "cust-val", int32(5000), true,
		"orders queue", string(QueueDefaultPersistenceYes), string(QueueAccessEnabledEnabled),
		string(QueueAccessEnabledDisabled), "TARGET.Q", "SYSTEM.XMIT", "QM2",
		true, string(DeletionPolicyDelete), string(AdoptionPolicyAdopt),
		"Available", "synced", "ok", "DEFINE QLOCAL(APP.MAX)")

	f.Fuzz(func(t *testing.T,
		name, namespace, uid, resourceVersion string, generation int64, populateMeta bool,
		connRef, queueName, queueType, attrVal string, maxDepth int32, setMaxDepth bool,
		description, defPersistence, get, put, targetQueue, xmitQueue, remoteQM string,
		suspend bool, deletionPolicy, adoptionPolicy string,
		condReason, condMessage, statusMessage, desiredMQSC string,
	) {
		orig := &Queue{
			ObjectMeta: fuzzMeta(name, namespace, uid, resourceVersion, generation, populateMeta),
			Spec: QueueSpec{
				ConnectionRef:             LocalObjectReference{Name: connRef},
				QueueName:                 queueName,
				Type:                      QueueType(queueType),
				Attributes:                fuzzAttrs(attrVal != "", attrVal),
				MaxDepth:                  fuzzInt32Ptr(setMaxDepth, maxDepth),
				Description:               description,
				DefPersistence:            QueueDefaultPersistence(defPersistence),
				Get:                       QueueAccessEnabled(get),
				Put:                       QueueAccessEnabled(put),
				TargetQueue:               targetQueue,
				XmitQueue:                 xmitQueue,
				RemoteQueueManager:        remoteQM,
				Suspend:                   suspend,
				WorkloadLifecyclePolicies: fuzzWorkloadPolicies(deletionPolicy, adoptionPolicy),
			},
			Status: QueueStatus{
				Conditions:           fuzzConditions(condReason != "", condReason, condMessage),
				ObservedGeneration:   generation,
				DesiredMQSC:          desiredMQSC,
				MQObjectStatusFields: fuzzMQStatus(statusMessage != "", statusMessage),
			},
		}
		_, back := roundTripQueue(t, orig.DeepCopy())
		assertLossless(t, orig, back)
	})
}

// FuzzTopicConversionRoundTrip fuzzes Topic ConvertTo/ConvertFrom for panics and loss.
func FuzzTopicConversionRoundTrip(f *testing.F) {
	f.Add("t-min", "default", "", "", int64(0), false,
		"qm1", "APP.MIN", "", "", "", "", "", "", "", "", false, "", "", "", "", "", "")
	f.Add("t-max", "prod", "uid-1", "42", int64(3), true,
		"qm1", "RETAIL.ORDERS", "cust-val", "orders/#", "topic desc",
		string(TopicAccessEnabledEnabled), string(TopicAccessEnabledDisabled),
		string(QueueDefaultPersistenceYes), "ALL", "QMGR",
		true, string(DeletionPolicyDelete), string(AdoptionPolicyAdopt),
		"Available", "synced", "ok", "DEFINE TOPIC(RETAIL.ORDERS)")

	f.Fuzz(func(t *testing.T,
		name, namespace, uid, resourceVersion string, generation int64, populateMeta bool,
		connRef, topicName, attrVal, topicString, description, publish, subscribe,
		defPersistence, publishScope, subscribeScope string,
		suspend bool, deletionPolicy, adoptionPolicy string,
		condReason, condMessage, statusMessage, desiredMQSC string,
	) {
		orig := &Topic{
			ObjectMeta: fuzzMeta(name, namespace, uid, resourceVersion, generation, populateMeta),
			Spec: TopicSpec{
				ConnectionRef:             LocalObjectReference{Name: connRef},
				TopicName:                 topicName,
				Attributes:                fuzzAttrs(attrVal != "", attrVal),
				TopicString:               topicString,
				Description:               description,
				Publish:                   TopicAccessEnabled(publish),
				Subscribe:                 TopicAccessEnabled(subscribe),
				DefPersistence:            QueueDefaultPersistence(defPersistence),
				PublishScope:              publishScope,
				SubscribeScope:            subscribeScope,
				Suspend:                   suspend,
				WorkloadLifecyclePolicies: fuzzWorkloadPolicies(deletionPolicy, adoptionPolicy),
			},
			Status: TopicStatus{
				Conditions:           fuzzConditions(condReason != "", condReason, condMessage),
				ObservedGeneration:   generation,
				DesiredMQSC:          desiredMQSC,
				MQObjectStatusFields: fuzzMQStatus(statusMessage != "", statusMessage),
			},
		}
		_, back := roundTripTopic(t, orig.DeepCopy())
		assertLossless(t, orig, back)
	})
}

// FuzzChannelConversionRoundTrip fuzzes Channel ConvertTo/ConvertFrom for panics and loss.
func FuzzChannelConversionRoundTrip(f *testing.F) {
	f.Add("c-min", "default", "", "", int64(0), false,
		"qm1", "APP.MIN", "", "", "", int32(0), false, "", int32(0), false, "",
		int32(0), false, int32(0), false, "", "", "", "",
		false, "", "", "", "", "", "")
	f.Add("c-max", "prod", "uid-1", "42", int64(3), true,
		"qm1", "ORDERS.APP", string(ChannelTypeSdr), "cust-val", "chan desc",
		int32(4194304), true, string(ChannelTransportTypeTCP), int32(10), true, "app",
		int32(50), true, int32(5), true, "TLS_RSA_WITH_AES_128_CBC_SHA256",
		string(ChannelSslClientAuthRequired), "qm2.example.com(1414)", "SYSTEM.XMIT",
		true, string(DeletionPolicyDelete), string(AdoptionPolicyAdopt),
		"Available", "synced", "ok", "DEFINE CHANNEL(ORDERS.APP)")

	f.Fuzz(func(t *testing.T,
		name, namespace, uid, resourceVersion string, generation int64, populateMeta bool,
		connRef, channelName, channelType, attrVal, description string,
		maxMsgLength int32, setMaxMsgLength bool, transportType string,
		shareConv int32, setShareConv bool, mcaUser string,
		maxInstances int32, setMaxInstances bool, maxInstancesClient int32, setMaxInstancesClient bool,
		sslCipherSpec, sslClientAuth, connName, xmitQueue string,
		suspend bool, deletionPolicy, adoptionPolicy string,
		condReason, condMessage, statusMessage, desiredMQSC string,
	) {
		orig := &Channel{
			ObjectMeta: fuzzMeta(name, namespace, uid, resourceVersion, generation, populateMeta),
			Spec: ChannelSpec{
				ConnectionRef:             LocalObjectReference{Name: connRef},
				ChannelName:               channelName,
				Type:                      ChannelType(channelType),
				Attributes:                fuzzAttrs(attrVal != "", attrVal),
				Description:               description,
				MaxMsgLength:              fuzzInt32Ptr(setMaxMsgLength, maxMsgLength),
				TransportType:             ChannelTransportType(transportType),
				ShareConv:                 fuzzInt32Ptr(setShareConv, shareConv),
				McaUser:                   mcaUser,
				MaxInstances:              fuzzInt32Ptr(setMaxInstances, maxInstances),
				MaxInstancesClient:        fuzzInt32Ptr(setMaxInstancesClient, maxInstancesClient),
				SslCipherSpec:             sslCipherSpec,
				SslClientAuth:             ChannelSslClientAuth(sslClientAuth),
				ConnName:                  connName,
				XmitQueue:                 xmitQueue,
				Suspend:                   suspend,
				WorkloadLifecyclePolicies: fuzzWorkloadPolicies(deletionPolicy, adoptionPolicy),
			},
			Status: ChannelStatus{
				Conditions:           fuzzConditions(condReason != "", condReason, condMessage),
				ObservedGeneration:   generation,
				DesiredMQSC:          desiredMQSC,
				MQObjectStatusFields: fuzzMQStatus(statusMessage != "", statusMessage),
			},
		}
		_, back := roundTripChannel(t, orig.DeepCopy())
		assertLossless(t, orig, back)
	})
}

// FuzzChannelAuthRuleConversionRoundTrip fuzzes ChannelAuthRule conversion.
func FuzzChannelAuthRuleConversionRoundTrip(f *testing.F) {
	f.Add("car-min", "default", "", "", int64(0), false,
		"qm1", "APP.SVRCONN", string(ChannelAuthRuleTypeBlockUser),
		"", "", "", "", "", "", "", "", "",
		false, "", "", "", "", "", "")
	f.Add("car-max", "prod", "uid-1", "42", int64(3), true,
		"qm1", "APP.SVRCONN", string(ChannelAuthRuleTypeAddressMap),
		"192.168.0.0/24", "baduser", "clientuser", "CN=app", "QM2", "app",
		string(ChannelAuthUserSourceMap), string(ChannelAuthCheckClientRequired), "chlauth desc",
		true, string(DeletionPolicyDelete), string(AdoptionPolicyAdopt),
		"Available", "synced", "ok", "SET CHLAUTH(APP.SVRCONN)")

	f.Fuzz(func(t *testing.T,
		name, namespace, uid, resourceVersion string, generation int64, populateMeta bool,
		connRef, channelName, ruleType, address, userList, clientUser, sslPeerName,
		remoteQM, mcaUser, userSource, checkClient, description string,
		suspend bool, deletionPolicy, adoptionPolicy string,
		condReason, condMessage, statusMessage, desiredMQSC string,
	) {
		orig := &ChannelAuthRule{
			ObjectMeta: fuzzMeta(name, namespace, uid, resourceVersion, generation, populateMeta),
			Spec: ChannelAuthRuleSpec{
				ConnectionRef:             LocalObjectReference{Name: connRef},
				ChannelName:               channelName,
				RuleType:                  ChannelAuthRuleType(ruleType),
				Address:                   address,
				UserList:                  userList,
				ClientUser:                clientUser,
				SslPeerName:               sslPeerName,
				RemoteQueueManager:        remoteQM,
				McaUser:                   mcaUser,
				UserSource:                ChannelAuthUserSource(userSource),
				CheckClient:               ChannelAuthCheckClient(checkClient),
				Description:               description,
				Suspend:                   suspend,
				WorkloadLifecyclePolicies: fuzzWorkloadPolicies(deletionPolicy, adoptionPolicy),
			},
			Status: ChannelAuthRuleStatus{
				Conditions:           fuzzConditions(condReason != "", condReason, condMessage),
				ObservedGeneration:   generation,
				DesiredMQSC:          desiredMQSC,
				MQObjectStatusFields: fuzzMQStatus(statusMessage != "", statusMessage),
			},
		}
		_, back := roundTripChannelAuthRule(t, orig.DeepCopy())
		assertLossless(t, orig, back)
	})
}

// FuzzAuthorityRecordConversionRoundTrip fuzzes AuthorityRecord conversion.
func FuzzAuthorityRecordConversionRoundTrip(f *testing.F) {
	f.Add("ar-min", "default", "", "", int64(0), false,
		"qm1", "APP.ORDERS", string(AuthorityObjectTypeQueue), "", "", "",
		false, "", "", "", "", "", "")
	f.Add("ar-max", "prod", "uid-1", "42", int64(3), true,
		"qm1", "APP.ORDERS", string(AuthorityObjectTypeChannel), "app", "apps", "GET,PUT,CONNECT",
		true, string(DeletionPolicyDelete), string(AdoptionPolicyAdopt),
		"Available", "synced", "ok", "SET AUTHREC(APP.ORDERS)")

	f.Fuzz(func(t *testing.T,
		name, namespace, uid, resourceVersion string, generation int64, populateMeta bool,
		connRef, profile, objectType, principal, group, authoritiesCSV string,
		suspend bool, deletionPolicy, adoptionPolicy string,
		condReason, condMessage, statusMessage, desiredMQSC string,
	) {
		orig := &AuthorityRecord{
			ObjectMeta: fuzzMeta(name, namespace, uid, resourceVersion, generation, populateMeta),
			Spec: AuthorityRecordSpec{
				ConnectionRef:             LocalObjectReference{Name: connRef},
				Profile:                   profile,
				ObjectType:                AuthorityObjectType(objectType),
				Principal:                 principal,
				Group:                     group,
				Authorities:               fuzzAuthorities(authoritiesCSV),
				Suspend:                   suspend,
				WorkloadLifecyclePolicies: fuzzWorkloadPolicies(deletionPolicy, adoptionPolicy),
			},
			Status: AuthorityRecordStatus{
				Conditions:           fuzzConditions(condReason != "", condReason, condMessage),
				ObservedGeneration:   generation,
				DesiredMQSC:          desiredMQSC,
				MQObjectStatusFields: fuzzMQStatus(statusMessage != "", statusMessage),
			},
		}
		_, back := roundTripAuthorityRecord(t, orig.DeepCopy())
		assertLossless(t, orig, back)
	})
}

// FuzzQueueManagerConnectionConversionRoundTrip fuzzes QueueManagerConnection conversion.
func FuzzQueueManagerConnectionConversionRoundTrip(f *testing.F) {
	// QueueManagerConnectionStatus has no DesiredMQSC field (unlike the other five
	// kinds), so the seed and signature omit it.
	f.Add("qmc-min", "default", "", "", int64(0), false,
		"QM1", "https://mq.svc:9443", "", false, false, "", "creds",
		"", "")
	f.Add("qmc-max", "prod", "uid-1", "42", int64(3), true,
		"QM1", "https://mq.svc:9443", "/ibmmq/rest/v3", true, true, "ca-secret", "creds",
		"Available", "synced")

	f.Fuzz(func(t *testing.T,
		name, namespace, uid, resourceVersion string, generation int64, populateMeta bool,
		queueManager, endpoint, restPrefix string, setTLS, insecureSkipVerify bool,
		caSecretRef, credsSecretRef string,
		condReason, condMessage string,
	) {
		orig := &QueueManagerConnection{
			ObjectMeta: fuzzMeta(name, namespace, uid, resourceVersion, generation, populateMeta),
			Spec: QueueManagerConnectionSpec{
				QueueManager:         queueManager,
				Endpoint:             endpoint,
				RESTPrefix:           restPrefix,
				TLS:                  fuzzTLSConfig(setTLS, insecureSkipVerify, caSecretRef),
				CredentialsSecretRef: SecretReference{Name: credsSecretRef},
			},
			Status: QueueManagerConnectionStatus{
				Conditions:         fuzzConditions(condReason != "", condReason, condMessage),
				ObservedGeneration: generation,
			},
		}
		_, back := roundTripQueueManagerConnection(t, orig.DeepCopy())
		assertLossless(t, orig, back)
	})
}

// fuzzAuthorities splits a CSV of authorities into a slice, returning nil when empty so
// it matches copyStringSlice's empty->nil normalization on the round-trip.
func fuzzAuthorities(csv string) []string {
	if csv == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i <= len(csv); i++ {
		if i == len(csv) || csv[i] == ',' {
			if i > start {
				out = append(out, csv[start:i])
			}
			start = i + 1
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// fuzzTLSConfig builds a TLSConfig when setTLS is set (with an optional CASecretRef),
// else nil. The converter branches on TLS==nil, so both cases exercise a distinct path.
func fuzzTLSConfig(setTLS, insecureSkipVerify bool, caSecretRef string) *TLSConfig {
	if !setTLS {
		return nil
	}
	cfg := &TLSConfig{InsecureSkipVerify: insecureSkipVerify}
	if caSecretRef != "" {
		cfg.CASecretRef = &SecretReference{Name: caSecretRef}
	}
	return cfg
}
