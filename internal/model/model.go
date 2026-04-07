package model

import "time"

const (
	DurableResourceName = "memos.notes.kubememo.io"
	RuntimeResourceName = "runtimememos.runtime.notes.kubememo.io"
	DurableAPIVersion   = "notes.kubememo.io/v1alpha1"
	RuntimeAPIVersion   = "runtime.notes.kubememo.io/v1alpha1"
	DurableKind         = "Memo"
	RuntimeKind         = "RuntimeMemo"
)

type Note struct {
	ID              string              `json:"id"`
	StoreType       string              `json:"storeType"`
	Title           string              `json:"title"`
	Summary         string              `json:"summary,omitempty"`
	Content         string              `json:"content,omitempty"`
	Format          string              `json:"format,omitempty"`
	NoteType        string              `json:"noteType"`
	Temporary       bool                `json:"temporary"`
	Severity        string              `json:"severity,omitempty"`
	OwnerTeam       string              `json:"ownerTeam,omitempty"`
	OwnerContact    string              `json:"ownerContact,omitempty"`
	Tags            []string            `json:"tags,omitempty"`
	Links           []map[string]string `json:"links,omitempty"`
	TargetMode      string              `json:"targetMode"`
	APIVersion      string              `json:"apiVersion,omitempty"`
	Kind            string              `json:"kind,omitempty"`
	Namespace       string              `json:"namespace,omitempty"`
	Name            string              `json:"name,omitempty"`
	AppName         string              `json:"appName,omitempty"`
	AppInstance     string              `json:"appInstance,omitempty"`
	ValidFrom       *time.Time          `json:"validFrom,omitempty"`
	ExpiresAt       *time.Time          `json:"expiresAt,omitempty"`
	CreatedAt       *time.Time          `json:"createdAt,omitempty"`
	UpdatedAt       *time.Time          `json:"updatedAt,omitempty"`
	CreatedBy       string              `json:"createdBy,omitempty"`
	UpdatedBy       string              `json:"updatedBy,omitempty"`
	SourceType      string              `json:"sourceType,omitempty"`
	SourceGenerator string              `json:"sourceGenerator,omitempty"`
	Confidence      string              `json:"confidence,omitempty"`
	GitRepo         string              `json:"gitRepo,omitempty"`
	GitPath         string              `json:"gitPath,omitempty"`
	GitRevision     string              `json:"gitRevision,omitempty"`
	Activity        map[string]any      `json:"activity,omitempty"`
	RawResource     map[string]any      `json:"rawResource,omitempty"`
}

type NoteList struct {
	Items []Note `json:"items"`
}

type InstallationStatus struct {
	ClusterReachable          bool              `json:"clusterReachable"`
	DurableCrdInstalled       bool              `json:"durableCrdInstalled"`
	RuntimeCrdInstalled       bool              `json:"runtimeCrdInstalled"`
	RuntimeNamespaceInstalled bool              `json:"runtimeNamespaceInstalled"`
	RbacInstalled             bool              `json:"rbacInstalled"`
	ActivityCaptureInstalled  bool              `json:"activityCaptureInstalled"`
	GitOps                    GitOpsState       `json:"gitOps"`
	RuntimeStore              RuntimeStoreState `json:"runtimeStore"`
}

type InstallationModeStatus struct {
	Mode   string             `json:"mode"`
	Status InstallationStatus `json:"status"`
}

type GitOpsState struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type RuntimeStoreState struct {
	Enabled bool   `json:"enabled"`
	Safe    bool   `json:"safe"`
	Reason  string `json:"reason,omitempty"`
}

type Target struct {
	Mode        string `json:"mode"`
	APIVersion  string `json:"apiVersion,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Name        string `json:"name,omitempty"`
	AppName     string `json:"appName,omitempty"`
	AppInstance string `json:"appInstance,omitempty"`
}
