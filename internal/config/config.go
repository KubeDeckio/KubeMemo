package config

type Config struct {
	Enabled          bool            `json:"enabled"`
	RuntimeNamespace string          `json:"runtimeNamespace"`
	GitOpsRepoPath   string          `json:"gitOpsRepoPath"`
	Storage          StorageConfig   `json:"storage"`
	Install          InstallConfig   `json:"install"`
	GitOps           GitOpsConfig    `json:"gitOps"`
	Runtime          RuntimeConfig   `json:"runtime"`
	ActivityCapture  ActivityConfig  `json:"activityCapture"`
	Rendering        RenderingConfig `json:"rendering"`
}

type StorageConfig struct {
	Durable StoreConfig `json:"durable"`
	Runtime StoreConfig `json:"runtime"`
}

type StoreConfig struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
	Kind    string `json:"kind"`
	Group   string `json:"group"`
}

type InstallConfig struct {
	ManageCrds             bool   `json:"manageCrds"`
	ManageRuntimeNamespace bool   `json:"manageRuntimeNamespace"`
	ManageRbac             string `json:"manageRbac"`
}

type GitOpsConfig struct {
	Enabled                 bool   `json:"enabled"`
	Provider                string `json:"provider"`
	DurableSourceOfTruth    string `json:"durableSourceOfTruth"`
	RepoPath                string `json:"repoPath"`
	AllowDirectDurableWrite bool   `json:"allowDirectDurableWrite"`
}

type RuntimeConfig struct {
	DefaultExpiryHours         int  `json:"defaultExpiryHours"`
	IncidentDefaultExpiryHours int  `json:"incidentDefaultExpiryHours"`
	AutoDeleteExpired          bool `json:"autoDeleteExpired"`
}

type ActivityConfig struct {
	Enabled             bool     `json:"enabled"`
	RequireNotesEnabled bool     `json:"requireNotesEnabled"`
	WriteTarget         string   `json:"writeTarget"`
	WatchKinds          []string `json:"watchKinds"`
	CaptureActions      []string `json:"captureActions"`
	DedupeWindowSeconds int      `json:"dedupeWindowSeconds"`
	Image               string   `json:"image"`
}

type RenderingConfig struct {
	Cards    bool `json:"cards"`
	Markdown bool `json:"markdown"`
	ANSI     bool `json:"ansi"`
}

func Default() Config {
	return Config{
		Enabled:          true,
		RuntimeNamespace: "kubememo-runtime",
		GitOpsRepoPath:   "./ops/kubememo",
		Storage: StorageConfig{
			Durable: StoreConfig{
				Enabled: true,
				Mode:    "crd",
				Kind:    "Memo",
				Group:   "notes.kubememo.io",
			},
			Runtime: StoreConfig{
				Enabled: true,
				Mode:    "crd",
				Kind:    "RuntimeMemo",
				Group:   "runtime.notes.kubememo.io",
			},
		},
		Install: InstallConfig{
			ManageCrds:             true,
			ManageRuntimeNamespace: true,
			ManageRbac:             "optional",
		},
		GitOps: GitOpsConfig{
			Enabled:                 false,
			Provider:                "auto",
			DurableSourceOfTruth:    "git",
			RepoPath:                "./ops/kubememo",
			AllowDirectDurableWrite: false,
		},
		Runtime: RuntimeConfig{
			DefaultExpiryHours:         24,
			IncidentDefaultExpiryHours: 12,
			AutoDeleteExpired:          false,
		},
		ActivityCapture: ActivityConfig{
			Enabled:             true,
			RequireNotesEnabled: true,
			WriteTarget:         "runtime",
			WatchKinds: []string{
				"Deployment",
				"StatefulSet",
				"DaemonSet",
				"Service",
				"Ingress",
				"HorizontalPodAutoscaler",
				"Namespace",
				"Gateway",
				"HTTPRoute",
			},
			CaptureActions: []string{
				"scale",
				"imageChange",
				"resourceChange",
				"serviceTypeChange",
				"ingressChange",
			},
			DedupeWindowSeconds: 60,
			Image:               "ghcr.io/kubedeckio/kubememo:latest",
		},
		Rendering: RenderingConfig{
			Cards:    true,
			Markdown: true,
			ANSI:     true,
		},
	}
}
