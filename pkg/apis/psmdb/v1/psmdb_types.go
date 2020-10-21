package v1

import (
	"encoding/json"
	"strings"

	"github.com/percona/percona-backup-mongodb/pbm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sversion "k8s.io/apimachinery/pkg/version"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	v "github.com/hashicorp/go-version"
	"github.com/percona/percona-server-mongodb-operator/version"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PerconaServerMongoDB is the Schema for the perconaservermongodbs API
// +k8s:openapi-gen=true
type PerconaServerMongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PerconaServerMongoDBSpec   `json:"spec,omitempty"`
	Status PerconaServerMongoDBStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PerconaServerMongoDBList contains a list of PerconaServerMongoDB
type PerconaServerMongoDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PerconaServerMongoDB `json:"items"`
}

type ClusterRole string

const (
	ClusterRoleShardSvr  ClusterRole = "shardsvr"
	ClusterRoleConfigSvr ClusterRole = "configsvr"
)

// PerconaServerMongoDBSpec defines the desired state of PerconaServerMongoDB
type PerconaServerMongoDBSpec struct {
	Pause                   bool                                 `json:"pause,omitempty"`
	CRVersion               string                               `json:"crVersion,omitempty"`
	Platform                *version.Platform                    `json:"platform,omitempty"`
	Image                   string                               `json:"image,omitempty"`
	ImagePullSecrets        []corev1.LocalObjectReference        `json:"imagePullSecrets,omitempty"`
	RunUID                  int64                                `json:"runUid,omitempty"`
	UnsafeConf              bool                                 `json:"allowUnsafeConfigurations"`
	Mongod                  *MongodSpec                          `json:"mongod,omitempty"`
	Replsets                []*ReplsetSpec                       `json:"replsets,omitempty"`
	Secrets                 *SecretsSpec                         `json:"secrets,omitempty"`
	Backup                  BackupSpec                           `json:"backup,omitempty"`
	ImagePullPolicy         corev1.PullPolicy                    `json:"imagePullPolicy,omitempty"`
	PMM                     PMMSpec                              `json:"pmm,omitempty"`
	UpdateStrategy          appsv1.StatefulSetUpdateStrategyType `json:"updateStrategy,omitempty"`
	UpgradeOptions          UpgradeOptions                       `json:"upgradeOptions,omitempty"`
	SchedulerName           string                               `json:"schedulerName,omitempty"`
	ClusterServiceDNSSuffix string                               `json:"clusterServiceDNSSuffix,omitempty"`
}

const (
	SmartUpdateStatefulSetStrategyType appsv1.StatefulSetUpdateStrategyType = "SmartUpdate"
)

type UpgradeOptions struct {
	VersionServiceEndpoint string          `json:"versionServiceEndpoint,omitempty"`
	Apply                  UpgradeStrategy `json:"apply,omitempty"`
	Schedule               string          `json:"schedule,omitempty"`
}

type ReplsetMemberStatus struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type ReplsetStatus struct {
	Members     []*ReplsetMemberStatus `json:"members,omitempty"`
	ClusterRole ClusterRole            `json:"clusterRole,omitempty"`

	Initialized bool     `json:"initialized,omitempty"`
	Size        int32    `json:"size"`
	Ready       int32    `json:"ready"`
	Status      AppState `json:"status,omitempty"`
	Message     string   `json:"message,omitempty"`
}

type AppState string

const (
	AppStatePending AppState = "pending"
	AppStateInit    AppState = "initializing"
	AppStateReady   AppState = "ready"
	AppStateError   AppState = "error"
)

type UpgradeStrategy string

func (us UpgradeStrategy) Lower() UpgradeStrategy {
	return UpgradeStrategy(strings.ToLower(string(us)))
}

const (
	UpgradeStrategyDiasbled UpgradeStrategy = "disabled"
	UpgradeStrategyNever    UpgradeStrategy = "never"
)

// PerconaServerMongoDBStatus defines the observed state of PerconaServerMongoDB
type PerconaServerMongoDBStatus struct {
	State              AppState                  `json:"state,omitempty"`
	MongoVersion       string                    `json:"mongoVersion,omitempty"`
	MongoImage         string                    `json:"mongoImage,omitempty"`
	Message            string                    `json:"message,omitempty"`
	Conditions         []ClusterCondition        `json:"conditions,omitempty"`
	Replsets           map[string]*ReplsetStatus `json:"replsets,omitempty"`
	ObservedGeneration int64                     `json:"observedGeneration,omitempty"`
	BackupStatus       AppState                  `json:"backup,omitempty"`
	BackupVersion      string                    `json:"backupVersion,omitempty"`
	PMMStatus          AppState                  `json:"pmmStatus,omitempty"`
	PMMVersion         string                    `json:"pmmVersion,omitempty"`
}

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type ClusterConditionType string

const (
	ClusterReady   ClusterConditionType = "ClusterReady"
	ClusterInit    ClusterConditionType = "ClusterInitializing"
	ClusterRSInit  ClusterConditionType = "ReplsetInitialized"
	ClusterRSReady ClusterConditionType = "ReplsetReady"
	ClusterError   ClusterConditionType = "Error"
)

type Mapping struct {
	External string `json:"external"`
	Internal string `json:"internal"`
}

type Connectivity struct {
	Mapping []Mapping `json:"mapping"`
}

type ClusterCondition struct {
	Status             ConditionStatus      `json:"status"`
	Type               ClusterConditionType `json:"type"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

type PMMSpec struct {
	Enabled    bool           `json:"enabled,omitempty"`
	ServerHost string         `json:"serverHost,omitempty"`
	Image      string         `json:"image,omitempty"`
	Resources  *ResourcesSpec `json:"resources,omitempty"`
}

type MultiAZ struct {
	Affinity            *PodAffinity             `json:"affinity,omitempty"`
	NodeSelector        map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations         []corev1.Toleration      `json:"tolerations,omitempty"`
	PriorityClassName   string                   `json:"priorityClassName,omitempty"`
	Annotations         map[string]string        `json:"annotations,omitempty"`
	Labels              map[string]string        `json:"labels,omitempty"`
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	MinAvailable   *intstr.IntOrString `json:"minAvailable,omitempty"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

type PodAffinity struct {
	TopologyKey *string          `json:"antiAffinityTopologyKey,omitempty"`
	Advanced    *corev1.Affinity `json:"advanced,omitempty"`
}

type ReplsetSpec struct {
	Resources                *ResourcesSpec             `json:"resources,omitempty"`
	Name                     string                     `json:"name"`
	Size                     int32                      `json:"size"`
	ClusterRole              ClusterRole                `json:"clusterRole,omitempty"`
	Arbiter                  Arbiter                    `json:"arbiter,omitempty"`
	Expose                   Expose                     `json:"expose,omitempty"`
	VolumeSpec               *VolumeSpec                `json:"volumeSpec,omitempty"`
	ReadinessProbe           *corev1.Probe              `json:"readinessProbe,omitempty"`
	LivenessProbe            *LivenessProbeExtended     `json:"livenessProbe,omitempty"`
	PodSecurityContext       *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	ContainerSecurityContext *corev1.SecurityContext    `json:"containerSecurityContext,omitempty"`
	Connectivity             *Connectivity              `json:"connectivity"`
	MultiAZ
}

type LivenessProbeExtended struct {
	corev1.Probe        `json:",inline"`
	StartupDelaySeconds int `json:"startupDelaySeconds,omitempty"`
}

func (l LivenessProbeExtended) CommandHas(flag string) bool {
	if l.Handler.Exec == nil {
		return false
	}

	for _, v := range l.Handler.Exec.Command {
		if v == flag {
			return true
		}
	}

	return false
}

type VolumeSpec struct {
	// EmptyDir represents a temporary directory that shares a pod's lifetime.
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// HostPath represents a pre-existing file or directory on the host machine
	// that is directly exposed to the container.
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// PersistentVolumeClaim represents a reference to a PersistentVolumeClaim.
	// It has the highest level of precedence, followed by HostPath and
	// EmptyDir. And represents the PVC specification.
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

type ResourceSpecRequirements struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type ResourcesSpec struct {
	Limits   *ResourceSpecRequirements `json:"limits,omitempty"`
	Requests *ResourceSpecRequirements `json:"requests,omitempty"`
}

type Issuer struct {
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type SecretsSpec struct {
	Users          string  `json:"users,omitempty"`
	SSL            string  `json:"ssl,omitempty"`
	SSLInternal    string  `json:"sslInternal,omitempty"`
	ExistingIssuer *Issuer `json:"existingIssuer,omitempty"`
}

type MongosSpec struct {
	*ResourcesSpec `json:"resources,omitempty"`
	Port           int32 `json:"port,omitempty"`
	HostPort       int32 `json:"hostPort,omitempty"`
}

type MongodSpec struct {
	Net                      *MongodSpecNet                `json:"net,omitempty"`
	AuditLog                 *MongodSpecAuditLog           `json:"auditLog,omitempty"`
	OperationProfiling       *MongodSpecOperationProfiling `json:"operationProfiling,omitempty"`
	Replication              *MongodSpecReplication        `json:"replication,omitempty"`
	Security                 *MongodSpecSecurity           `json:"security,omitempty"`
	SetParameter             *MongodSpecSetParameter       `json:"setParameter,omitempty"`
	Storage                  *MongodSpecStorage            `json:"storage,omitempty"`
	LoadBalancerSourceRanges []string                      `json:"loadBalancerSourceRanges,omitempty"`
	ServiceAnnotations       map[string]string             `json:"serviceAnnotations,omitempty"`
}

type MongodSpecNet struct {
	Port     int32 `json:"port,omitempty"`
	HostPort int32 `json:"hostPort,omitempty"`
}

type MongodSpecReplication struct {
	OplogSizeMB int `json:"oplogSizeMB,omitempty"`
}

// MongodChiperMode is a cipher mode used by Data-at-Rest Encryption
type MongodChiperMode string

const (
	MongodChiperModeUnset MongodChiperMode = ""
	MongodChiperModeCBC   MongodChiperMode = "AES256-CBC"
	MongodChiperModeGCM   MongodChiperMode = "AES256-GCM"
)

type MongodSpecSecurity struct {
	RedactClientLogData  bool             `json:"redactClientLogData,omitempty"`
	EnableEncryption     *bool            `json:"enableEncryption,omitempty"`
	EncryptionKeySecret  string           `json:"encryptionKeySecret,omitempty"`
	EncryptionCipherMode MongodChiperMode `json:"encryptionCipherMode,omitempty"`
}

type MongodSpecSetParameter struct {
	TTLMonitorSleepSecs                   int `json:"ttlMonitorSleepSecs,omitempty"`
	WiredTigerConcurrentReadTransactions  int `json:"wiredTigerConcurrentReadTransactions,omitempty"`
	WiredTigerConcurrentWriteTransactions int `json:"wiredTigerConcurrentWriteTransactions,omitempty"`
	CursorTimeoutMillis                   int `json:"cursorTimeoutMillis,omitempty"`
}

type StorageEngine string

var (
	StorageEngineWiredTiger StorageEngine = "wiredTiger"
	StorageEngineInMemory   StorageEngine = "inMemory"
	StorageEngineMMAPv1     StorageEngine = "mmapv1"
)

type MongodSpecStorage struct {
	Engine         StorageEngine         `json:"engine,omitempty"`
	DirectoryPerDB bool                  `json:"directoryPerDB,omitempty"`
	SyncPeriodSecs int                   `json:"syncPeriodSecs,omitempty"`
	InMemory       *MongodSpecInMemory   `json:"inMemory,omitempty"`
	MMAPv1         *MongodSpecMMAPv1     `json:"mmapv1,omitempty"`
	WiredTiger     *MongodSpecWiredTiger `json:"wiredTiger,omitempty"`
}

type MongodSpecMMAPv1 struct {
	NsSize     int  `json:"nsSize,omitempty"`
	Smallfiles bool `json:"smallfiles,omitempty"`
}

type WiredTigerCompressor string

var (
	WiredTigerCompressorNone   WiredTigerCompressor = "none"
	WiredTigerCompressorSnappy WiredTigerCompressor = "snappy"
	WiredTigerCompressorZlib   WiredTigerCompressor = "zlib"
)

type MongodSpecWiredTigerEngineConfig struct {
	CacheSizeRatio      float64               `json:"cacheSizeRatio,omitempty"`
	DirectoryForIndexes bool                  `json:"directoryForIndexes,omitempty"`
	JournalCompressor   *WiredTigerCompressor `json:"journalCompressor,omitempty"`
}

type MongodSpecWiredTigerCollectionConfig struct {
	BlockCompressor *WiredTigerCompressor `json:"blockCompressor,omitempty"`
}

type MongodSpecWiredTigerIndexConfig struct {
	PrefixCompression bool `json:"prefixCompression,omitempty"`
}

type MongodSpecWiredTiger struct {
	CollectionConfig *MongodSpecWiredTigerCollectionConfig `json:"collectionConfig,omitempty"`
	EngineConfig     *MongodSpecWiredTigerEngineConfig     `json:"engineConfig,omitempty"`
	IndexConfig      *MongodSpecWiredTigerIndexConfig      `json:"indexConfig,omitempty"`
}

type MongodSpecInMemoryEngineConfig struct {
	InMemorySizeRatio float64 `json:"inMemorySizeRatio,omitempty"`
}

type MongodSpecInMemory struct {
	EngineConfig *MongodSpecInMemoryEngineConfig `json:"engineConfig,omitempty"`
}

type AuditLogDestination string

var AuditLogDestinationFile AuditLogDestination = "file"

type AuditLogFormat string

var (
	AuditLogFormatBSON AuditLogFormat = "BSON"
	AuditLogFormatJSON AuditLogFormat = "JSON"
)

type MongodSpecAuditLog struct {
	Destination AuditLogDestination `json:"destination,omitempty"`
	Format      AuditLogFormat      `json:"format,omitempty"`
	Filter      string              `json:"filter,omitempty"`
}

type OperationProfilingMode string

const (
	OperationProfilingModeAll    OperationProfilingMode = "all"
	OperationProfilingModeSlowOp OperationProfilingMode = "slowOp"
)

type MongodSpecOperationProfiling struct {
	Mode              OperationProfilingMode `json:"mode,omitempty"`
	SlowOpThresholdMs int                    `json:"slowOpThresholdMs,omitempty"`
	RateLimit         int                    `json:"rateLimit,omitempty"`
}

type BackupTaskSpec struct {
	Name            string              `json:"name"`
	Enabled         bool                `json:"enabled"`
	Schedule        string              `json:"schedule,omitempty"`
	StorageName     string              `json:"storageName,omitempty"`
	CompressionType pbm.CompressionType `json:"compressionType,omitempty"`
}

type BackupStorageS3Spec struct {
	Bucket            string `json:"bucket"`
	Prefix            string `json:"prefix,omitempty"`
	Region            string `json:"region,omitempty"`
	EndpointURL       string `json:"endpointUrl,omitempty"`
	CredentialsSecret string `json:"credentialsSecret"`
}

type BackupStorageType string

const (
	BackupStorageFilesystem BackupStorageType = "filesystem"
	BackupStorageS3         BackupStorageType = "s3"
)

type BackupStorageSpec struct {
	Type BackupStorageType   `json:"type"`
	S3   BackupStorageS3Spec `json:"s3,omitempty"`
}

type BackupSpec struct {
	Enabled                  bool                         `json:"enabled"`
	Storages                 map[string]BackupStorageSpec `json:"storages,omitempty"`
	Image                    string                       `json:"image,omitempty"`
	Tasks                    []BackupTaskSpec             `json:"tasks,omitempty"`
	ServiceAccountName       string                       `json:"serviceAccountName,omitempty"`
	PodSecurityContext       *corev1.PodSecurityContext   `json:"podSecurityContext,omitempty"`
	ContainerSecurityContext *corev1.SecurityContext      `json:"containerSecurityContext,omitempty"`
	Resources                *ResourcesSpec               `json:"resources,omitempty"`
}

type Arbiter struct {
	Enabled bool  `json:"enabled"`
	Size    int32 `json:"size"`
	MultiAZ
}

type Expose struct {
	Enabled    bool               `json:"enabled"`
	ExposeType corev1.ServiceType `json:"exposeType,omitempty"`
}

// ServerVersion represents info about k8s / openshift server version
type ServerVersion struct {
	Platform version.Platform
	Info     k8sversion.Info
}

// OwnerRef returns OwnerReference to object
func (cr *PerconaServerMongoDB) OwnerRef(scheme *runtime.Scheme) (metav1.OwnerReference, error) {
	gvk, err := apiutil.GVKForObject(cr, scheme)
	if err != nil {
		return metav1.OwnerReference{}, err
	}

	trueVar := true

	return metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       cr.GetName(),
		UID:        cr.GetUID(),
		Controller: &trueVar,
	}, nil
}

// setVersion sets the API version of a PSMDB resource.
// The new (semver-matching) version is determined either by the CR's API version or an API version specified via the CR's annotations.
// If the CR's API version is an empty string, it returns "v1"
func (cr *PerconaServerMongoDB) setVersion() error {
	if len(cr.Spec.CRVersion) > 0 {
		return nil
	}

	apiVersion := version.Version

	if lastCR, ok := cr.Annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		var newCR PerconaServerMongoDB
		err := json.Unmarshal([]byte(lastCR), &newCR)
		if err != nil {
			return err
		}
		if len(newCR.APIVersion) > 0 {
			apiVersion = strings.Replace(strings.TrimPrefix(newCR.APIVersion, "psmdb.percona.com/v"), "-", ".", -1)
		}
	}

	cr.Spec.CRVersion = apiVersion

	return nil
}

func (cr *PerconaServerMongoDB) Version() *v.Version {
	return v.Must(v.NewVersion(cr.Spec.CRVersion))
}

func (cr *PerconaServerMongoDB) CompareVersion(version string) int {
	if len(cr.Spec.CRVersion) == 0 {
		cr.setVersion()
	}

	//using Must because "version" must be right format
	return cr.Version().Compare(v.Must(v.NewVersion(version)))
}
