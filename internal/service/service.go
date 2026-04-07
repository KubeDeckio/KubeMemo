package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/assets"
	"github.com/KubeDeckio/KubeMemo/internal/config"
	"github.com/KubeDeckio/KubeMemo/internal/kube"
	"github.com/KubeDeckio/KubeMemo/internal/model"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

var (
	durableGVR = schema.GroupVersionResource{Group: "notes.kubememo.io", Version: "v1alpha1", Resource: "memos"}
	runtimeGVR = schema.GroupVersionResource{Group: "runtime.notes.kubememo.io", Version: "v1alpha1", Resource: "runtimememos"}
)

const (
	annotationEnabled        = "notes.kubememo.io/enabled"
	annotationHasNote        = "notes.kubememo.io/has-note"
	annotationNoteCount      = "notes.kubememo.io/note-count"
	annotationRuntimeCount   = "notes.kubememo.io/runtime-count"
	annotationSummary        = "notes.kubememo.io/summary"
	annotationView           = "notes.kubememo.io/view"
	annotationRuntimeEnabled = "notes.kubememo.io/runtime-enabled"
	maxAnnotationSummaryLen  = 120
)

type Service struct {
	cfg  config.Config
	kube *kube.Client
}

type PersistResult struct {
	Note       model.Note `json:"note"`
	OutputPath string     `json:"outputPath,omitempty"`
	Manifest   string     `json:"manifest,omitempty"`
}

func New(cfg config.Config) (*Service, error) {
	client, err := kube.New()
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, kube: client}, nil
}

func NewWithClient(cfg config.Config, client *kube.Client) *Service {
	return &Service{cfg: cfg, kube: client}
}

func (s *Service) Config() config.Config {
	return s.cfg
}

func (s *Service) GetActor(ctx context.Context) string {
	return s.kube.GetActor(ctx)
}

func (s *Service) ResolveTarget(kind, namespace, name, apiVersion, targetNamespace, appName, appInstance string) model.Target {
	switch {
	case kind != "" && name != "":
		apiVersion = inferTargetAPIVersion(kind, apiVersion)
		return model.Target{
			Mode:        "resource",
			APIVersion:  apiVersion,
			Kind:        kind,
			Namespace:   namespace,
			Name:        name,
			AppName:     appName,
			AppInstance: appInstance,
		}
	case targetNamespace != "":
		return model.Target{
			Mode:      "namespace",
			Namespace: targetNamespace,
		}
	default:
		return model.Target{
			Mode:        "app",
			AppName:     appName,
			AppInstance: appInstance,
		}
	}
}

func inferTargetAPIVersion(kind, apiVersion string) string {
	if strings.TrimSpace(apiVersion) != "" && apiVersion != "v1" {
		return apiVersion
	}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "deployment", "statefulset", "daemonset", "replicaset":
		return "apps/v1"
	case "ingress":
		return "networking.k8s.io/v1"
	case "horizontalpodautoscaler":
		return "autoscaling/v2"
	case "gateway", "httproute":
		return "gateway.networking.k8s.io/v1"
	default:
		if strings.TrimSpace(apiVersion) != "" {
			return apiVersion
		}
		return "v1"
	}
}

func (s *Service) TestGitOpsMode(ctx context.Context) model.GitOpsState {
	namespaces := []string{"argocd", "flux-system"}
	for _, ns := range namespaces {
		_, err := s.kube.Clientset().CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err == nil {
			provider := "argocd"
			if ns == "flux-system" {
				provider = "flux"
			}
			return model.GitOpsState{Enabled: true, Provider: provider, Reason: fmt.Sprintf("detected namespace %s", ns)}
		}
	}
	return model.GitOpsState{Enabled: false, Provider: "none"}
}

func (s *Service) TestRuntimeStore(ctx context.Context, namespace string) model.RuntimeStoreState {
	if namespace == "" {
		namespace = s.cfg.RuntimeNamespace
	}
	_, err := s.kube.List(ctx, runtimeGVR, namespace)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return model.RuntimeStoreState{Enabled: false, Safe: false, Reason: "runtime CRD is not installed"}
		}
		if k8serrors.IsForbidden(err) {
			return model.RuntimeStoreState{Enabled: true, Safe: false, Reason: "runtime namespace is not readable with current RBAC"}
		}
		return model.RuntimeStoreState{Enabled: false, Safe: false, Reason: err.Error()}
	}
	gitOps := s.TestGitOpsMode(ctx)
	if gitOps.Enabled && (namespace == "argocd" || namespace == "flux-system") {
		return model.RuntimeStoreState{Enabled: true, Safe: false, Reason: "runtime namespace appears inside GitOps control plane namespaces"}
	}
	return model.RuntimeStoreState{Enabled: true, Safe: true}
}

func (s *Service) TestInstallation(ctx context.Context, runtimeNamespace string) model.InstallationStatus {
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	status := model.InstallationStatus{}
	if _, err := s.kube.ServerVersion(ctx); err == nil {
		status.ClusterReachable = true
	}
	_, durableErr := s.kube.List(ctx, durableGVR, "")
	status.DurableCrdInstalled = durableErr == nil || !k8serrors.IsNotFound(durableErr)
	_, runtimeErr := s.kube.List(ctx, runtimeGVR, runtimeNamespace)
	status.RuntimeCrdInstalled = runtimeErr == nil || !k8serrors.IsNotFound(runtimeErr)
	_, nsErr := s.kube.Clientset().CoreV1().Namespaces().Get(ctx, runtimeNamespace, metav1.GetOptions{})
	status.RuntimeNamespaceInstalled = nsErr == nil
	status.RbacInstalled = s.testRbac(ctx)
	status.ActivityCaptureInstalled = s.testActivityCapture(ctx, runtimeNamespace)
	status.Capabilities = s.capabilitySummary(ctx, runtimeNamespace)
	status.GitOps = s.TestGitOpsMode(ctx)
	status.RuntimeStore = s.TestRuntimeStore(ctx, runtimeNamespace)
	return status
}

func (s *Service) testRbac(ctx context.Context) bool {
	_, err := s.kube.Clientset().RbacV1().ClusterRoles().Get(ctx, "kubememo-reader", metav1.GetOptions{})
	return err == nil
}

func (s *Service) testActivityCapture(ctx context.Context, runtimeNamespace string) bool {
	_, err := s.kube.Clientset().AppsV1().Deployments(runtimeNamespace).Get(ctx, "kubememo-activity-capture", metav1.GetOptions{})
	return err == nil
}

func (s *Service) GetInstallationStatus(ctx context.Context, runtimeNamespace string) model.InstallationModeStatus {
	status := s.TestInstallation(ctx, runtimeNamespace)
	mode := "standard"
	if status.GitOps.Enabled && !status.RuntimeStore.Enabled {
		mode = "GitOps durable-only"
	} else if status.GitOps.Enabled && status.RuntimeStore.Enabled && status.RuntimeStore.Safe {
		mode = "GitOps with runtime store"
	}
	return model.InstallationModeStatus{Mode: mode, Status: status}
}

func (s *Service) capabilitySummary(ctx context.Context, runtimeNamespace string) model.CapabilitySummary {
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	currentNS := s.kube.CurrentNamespace()
	return model.CapabilitySummary{
		DurableRead:      s.checkCapability(ctx, "list", durableGVR, ""),
		DurableWrite:     s.checkCapability(ctx, "create", durableGVR, currentNS),
		RuntimeRead:      s.checkCapability(ctx, "list", runtimeGVR, runtimeNamespace),
		RuntimeWrite:     s.checkCapability(ctx, "create", runtimeGVR, runtimeNamespace),
		AnnotationPatch:  s.checkTargetPatchCapability(ctx),
		ActivityCapture:  s.checkActivityCaptureCapability(ctx, runtimeNamespace),
		ClusterScopeRead: s.checkCapability(ctx, "list", durableGVR, ""),
	}
}

func (s *Service) checkCapability(ctx context.Context, verb string, gvr schema.GroupVersionResource, namespace string) model.CapabilityState {
	allowed, reason := s.kube.CanI(ctx, verb, gvr, namespace)
	if allowed {
		scope := "cluster scope"
		if strings.TrimSpace(namespace) != "" {
			scope = "namespace " + namespace
		}
		return model.CapabilityState{Allowed: true, Reason: fmt.Sprintf("%s %s allowed on %s", verb, gvr.Resource, scope)}
	}
	if strings.TrimSpace(reason) == "" {
		reason = fmt.Sprintf("%s denied on %s", verb, gvr.Resource)
	}
	return model.CapabilityState{Allowed: false, Reason: reason}
}

func (s *Service) checkTargetPatchCapability(ctx context.Context) model.CapabilityState {
	targets := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "", Version: "v1", Resource: "namespaces"},
	}
	for _, gvr := range targets {
		ns := s.kube.CurrentNamespace()
		if gvr.Group == "" && gvr.Resource == "namespaces" {
			ns = ""
		}
		allowed, reason := s.kube.CanI(ctx, "patch", gvr, ns)
		if allowed {
			return model.CapabilityState{Allowed: true, Reason: fmt.Sprintf("resource annotation updates allowed on %s", gvr.Resource)}
		}
		if strings.TrimSpace(reason) != "" {
			return model.CapabilityState{Allowed: false, Reason: reason}
		}
	}
	return model.CapabilityState{Allowed: false, Reason: "resource annotation updates are not allowed with current RBAC"}
}

func (s *Service) checkActivityCaptureCapability(ctx context.Context, runtimeNamespace string) model.CapabilityState {
	kinds := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "gateways"},
	}
	for _, gvr := range kinds {
		allowed, reason := s.kube.CanI(ctx, "watch", gvr, "")
		if !allowed {
			if strings.TrimSpace(reason) == "" {
				reason = fmt.Sprintf("watch denied on %s", gvr.Resource)
			}
			return model.CapabilityState{Allowed: false, Reason: reason}
		}
	}
	runtimeWrite := s.checkCapability(ctx, "create", runtimeGVR, runtimeNamespace)
	if !runtimeWrite.Allowed {
		return model.CapabilityState{Allowed: false, Reason: "activity capture cannot write runtime memos: " + runtimeWrite.Reason}
	}
	return model.CapabilityState{Allowed: true, Reason: "watch and runtime memo write permissions are present"}
}

func (s *Service) Install(ctx context.Context, durableOnly, enableRuntimeStore bool, runtimeNamespace string, installRbac, gitOpsAware, enableActivityCapture bool, activityCaptureImage string) (model.InstallationModeStatus, error) {
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	if strings.TrimSpace(activityCaptureImage) == "" {
		activityCaptureImage = s.cfg.ActivityCapture.Image
	}
	gitOps := s.TestGitOpsMode(ctx)
	if gitOpsAware && gitOps.Enabled && !enableRuntimeStore {
		durableOnly = true
	}
	if enableActivityCapture {
		enableRuntimeStore = true
		installRbac = true
	}
	if err := s.kube.ApplyYAML(ctx, assets.DurableCRDYAML, runtimeNamespace); err != nil {
		return model.InstallationModeStatus{}, err
	}
	if !durableOnly || enableRuntimeStore {
		if err := s.kube.ApplyYAML(ctx, assets.RuntimeCRDYAML, runtimeNamespace); err != nil {
			return model.InstallationModeStatus{}, err
		}
		if err := s.kube.ApplyYAML(ctx, assets.RuntimeNamespaceYAML, runtimeNamespace); err != nil {
			return model.InstallationModeStatus{}, err
		}
	}
	if installRbac {
		if err := s.kube.ApplyYAML(ctx, assets.RBACYAML, runtimeNamespace); err != nil {
			return model.InstallationModeStatus{}, err
		}
	}
	if enableActivityCapture {
		if err := s.applyActivityCaptureDeployment(ctx, runtimeNamespace, activityCaptureImage); err != nil {
			return model.InstallationModeStatus{}, err
		}
	}
	status := s.GetInstallationStatus(ctx, runtimeNamespace)
	if gitOpsAware && gitOps.Enabled && enableRuntimeStore && !status.Status.RuntimeStore.Safe {
		return model.InstallationModeStatus{}, fmt.Errorf("runtime store is not safe for GitOps mode: %s", status.Status.RuntimeStore.Reason)
	}
	return status, nil
}

func (s *Service) Update(ctx context.Context, includeRbac bool, runtimeNamespace string, gitOpsAware, enableActivityCapture bool, activityCaptureImage string) (model.InstallationModeStatus, error) {
	return s.Install(ctx, false, true, runtimeNamespace, includeRbac, gitOpsAware, enableActivityCapture, activityCaptureImage)
}

func (s *Service) Uninstall(ctx context.Context, runtimeOnly, removeData bool, runtimeNamespace string) (map[string]any, error) {
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	result := map[string]any{
		"runtimeOnly": runtimeOnly,
		"removeData":  removeData,
	}
	_ = s.kube.Clientset().AppsV1().Deployments(runtimeNamespace).Delete(ctx, "kubememo-activity-capture", metav1.DeleteOptions{})
	if runtimeOnly {
		err := s.kube.Clientset().CoreV1().Namespaces().Delete(ctx, runtimeNamespace, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}
		result["removed"] = []string{runtimeNamespace}
		return result, nil
	}
	removed := []string{}
	if removeData {
		for _, gvr := range []schema.GroupVersionResource{runtimeGVR, durableGVR} {
			list, err := s.kube.List(ctx, gvr, "")
			if err == nil {
				for _, item := range list.Items {
					_ = s.kube.Delete(ctx, gvr, item.GetNamespace(), item.GetName())
				}
			}
		}
	}
	for _, gvr := range []schema.GroupVersionResource{runtimeGVR, durableGVR} {
		_ = s.kube.Delete(ctx, schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}, "", fmt.Sprintf("%s.%s", gvr.Resource, gvr.Group))
		removed = append(removed, fmt.Sprintf("%s.%s", gvr.Resource, gvr.Group))
	}
	err := s.kube.Clientset().CoreV1().Namespaces().Delete(ctx, runtimeNamespace, metav1.DeleteOptions{})
	if err == nil || k8serrors.IsNotFound(err) {
		removed = append(removed, runtimeNamespace)
	}
	result["removed"] = removed
	return result, nil
}

func (s *Service) applyActivityCaptureDeployment(ctx context.Context, runtimeNamespace, image string) error {
	deployment := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      "kubememo-activity-capture",
			"namespace": runtimeNamespace,
			"labels": map[string]any{
				"app.kubernetes.io/name":      "kubememo-activity-capture",
				"app.kubernetes.io/part-of":   "kubememo",
				"app.kubernetes.io/component": "activity-capture",
			},
		},
		"spec": map[string]any{
			"replicas": int64(1),
			"selector": map[string]any{
				"matchLabels": map[string]any{
					"app.kubernetes.io/name": "kubememo-activity-capture",
				},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app.kubernetes.io/name":      "kubememo-activity-capture",
						"app.kubernetes.io/part-of":   "kubememo",
						"app.kubernetes.io/component": "activity-capture",
					},
				},
				"spec": map[string]any{
					"serviceAccountName":           "kubememo",
					"automountServiceAccountToken": true,
					"securityContext": map[string]any{
						"runAsNonRoot": true,
						"runAsUser":    int64(65532),
						"runAsGroup":   int64(65532),
						"seccompProfile": map[string]any{
							"type": "RuntimeDefault",
						},
					},
					"containers": []any{
						map[string]any{
							"name":            "kubememo",
							"image":           image,
							"imagePullPolicy": "IfNotPresent",
							"args": []any{
								"watch-activity",
								"--runtime-namespace", runtimeNamespace,
							},
							"securityContext": map[string]any{
								"allowPrivilegeEscalation": false,
								"readOnlyRootFilesystem":   true,
								"capabilities": map[string]any{
									"drop": []any{"ALL"},
								},
							},
							"resources": map[string]any{
								"requests": map[string]any{
									"cpu":    "50m",
									"memory": "64Mi",
								},
								"limits": map[string]any{
									"cpu":    "250m",
									"memory": "256Mi",
								},
							},
							"env": []any{
								map[string]any{
									"name": "POD_NAMESPACE",
									"valueFrom": map[string]any{
										"fieldRef": map[string]any{
											"fieldPath": "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	return s.kube.ApplyUnstructured(ctx, deployment)
}

func (s *Service) ListNotes(ctx context.Context, includeRuntime bool, runtimeNamespace string, namespaces []string) ([]model.Note, error) {
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	results := []model.Note{}
	if len(namespaces) == 0 {
		list, err := s.kube.List(ctx, durableGVR, "")
		if err != nil {
			if k8serrors.IsForbidden(err) {
				list, err = s.kube.List(ctx, durableGVR, s.kube.CurrentNamespace())
			}
			if err != nil {
				return nil, permissionError(err, "read durable memos")
			}
		}
		for _, item := range list.Items {
			results = append(results, toNote(item, "Durable"))
		}
	} else {
		for _, ns := range namespaces {
			list, err := s.kube.List(ctx, durableGVR, ns)
			if err != nil {
				if k8serrors.IsForbidden(err) || k8serrors.IsNotFound(err) {
					continue
				}
				return nil, permissionError(err, "read durable memos")
			}
			for _, item := range list.Items {
				results = append(results, toNote(item, "Durable"))
			}
		}
	}
	if includeRuntime {
		list, err := s.kube.List(ctx, runtimeGVR, runtimeNamespace)
		if err == nil {
			for _, item := range list.Items {
				results = append(results, toNote(item, "Runtime"))
			}
		} else if !k8serrors.IsForbidden(err) && !k8serrors.IsNotFound(err) {
			return nil, permissionError(err, "read runtime memos")
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
	return results, nil
}

func (s *Service) FindNotes(ctx context.Context, text, noteType, kind, namespace, name string, includeRuntime, includeExpired bool, runtimeNamespace string) ([]model.Note, error) {
	namespaces := []string{}
	if namespace != "" {
		namespaces = []string{namespace}
	}
	notes, err := s.ListNotes(ctx, includeRuntime, runtimeNamespace, namespaces)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.Note, 0, len(notes))
	query := strings.ToLower(text)
	for _, note := range notes {
		if noteType != "" && !strings.EqualFold(note.NoteType, noteType) {
			continue
		}
		if kind != "" && !strings.EqualFold(note.Kind, kind) {
			continue
		}
		if namespace != "" && note.Namespace != namespace {
			continue
		}
		if name != "" && note.Name != name {
			continue
		}
		if !includeExpired && note.ExpiresAt != nil && note.ExpiresAt.Before(time.Now().UTC()) {
			continue
		}
		if query != "" {
			hay := strings.ToLower(strings.Join([]string{note.Title, note.Summary, note.Content, note.Name, note.Namespace, note.Kind}, " "))
			if !strings.Contains(hay, query) {
				continue
			}
		}
		filtered = append(filtered, note)
	}
	return filtered, nil
}

type NewNoteInput struct {
	Title            string
	Summary          string
	Content          string
	Format           string
	NoteType         string
	Severity         string
	OwnerTeam        string
	OwnerContact     string
	Tags             []string
	ExpiresAt        *time.Time
	Temporary        bool
	RuntimeNamespace string
	OutputPath       string
	AnnotateResource bool
	Target           model.Target
}

func (s *Service) NewNote(ctx context.Context, input NewNoteInput) (PersistResult, error) {
	resourceName := slugify(input.Title)
	if resourceName == "" {
		resourceName = fmt.Sprintf("kubememo-%d", time.Now().Unix())
	}
	now := time.Now().UTC()
	if input.Format == "" {
		input.Format = "markdown"
	}
	if input.NoteType == "" {
		input.NoteType = "advisory"
	}
	if input.Severity == "" {
		input.Severity = "info"
	}
	if input.RuntimeNamespace == "" {
		input.RuntimeNamespace = s.cfg.RuntimeNamespace
	}
	actor := s.GetActor(ctx)
	gitOps := s.TestGitOpsMode(ctx)
	obj := buildNoteObject(resourceName, input, actor, now, gitOps.Enabled)
	store := "Durable"
	gvr := durableGVR
	namespace := input.Target.Namespace
	if input.Target.Mode == "namespace" {
		namespace = input.Target.Namespace
	}
	if namespace == "" {
		namespace = s.kube.CurrentNamespace()
	}
	obj.SetNamespace(namespace)
	if input.Temporary {
		store = "Runtime"
		gvr = runtimeGVR
		obj.SetNamespace(input.RuntimeNamespace)
	}
	if !input.Temporary && (input.OutputPath != "" || gitOps.Enabled) {
		if input.OutputPath == "" {
			return PersistResult{}, fmt.Errorf("durable memo writes in GitOps mode require --output-path")
		}
		if input.AnnotateResource {
			return PersistResult{}, fmt.Errorf("resource annotation requires a direct cluster write and cannot be used with --output-path")
		}
		return s.writeManifestResult(obj, store, input.OutputPath)
	}
	if err := s.kube.ApplyUnstructured(ctx, obj); err != nil {
		return PersistResult{}, permissionError(err, "create "+strings.ToLower(store)+" memo")
	}
	created, err := s.kube.Dynamic().Resource(gvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		return PersistResult{}, err
	}
	note := toNote(*created, store)
	if s.shouldSyncAnnotations(noteToTarget(note), input.AnnotateResource) {
		if err := s.syncTargetAnnotations(ctx, noteToTarget(note), input.RuntimeNamespace, input.AnnotateResource, note.ID); err != nil {
			return PersistResult{}, err
		}
	}
	return PersistResult{Note: note}, nil
}

type UpdateNoteInput struct {
	ID               string
	Title            string
	Summary          *string
	Content          *string
	Tags             []string
	ExpiresAt        *time.Time
	Runtime          bool
	RuntimeNamespace string
	OutputPath       string
	AnnotateResource bool
}

func (s *Service) SetNote(ctx context.Context, input UpdateNoteInput) (PersistResult, error) {
	note, obj, gvr, ns, err := s.getNoteResource(ctx, input.ID, input.Runtime, input.RuntimeNamespace)
	if err != nil {
		return PersistResult{}, err
	}
	spec, _, _ := unstructured.NestedMap(obj.Object, "spec")
	if input.Title != "" {
		spec["title"] = input.Title
	}
	if input.Summary != nil {
		spec["summary"] = *input.Summary
	}
	if input.Content != nil {
		spec["content"] = *input.Content
	}
	if input.Tags != nil {
		tagValues := make([]any, 0, len(input.Tags))
		for _, tag := range input.Tags {
			tagValues = append(tagValues, tag)
		}
		spec["tags"] = tagValues
	}
	if input.ExpiresAt != nil {
		spec["expiresAt"] = input.ExpiresAt.UTC().Format(time.RFC3339)
	}
	spec["updatedBy"] = s.GetActor(ctx)
	_ = unstructured.SetNestedMap(obj.Object, spec, "spec")
	gitOps := s.TestGitOpsMode(ctx)
	if note.StoreType == "Durable" && (input.OutputPath != "" || gitOps.Enabled) {
		if input.OutputPath == "" {
			return PersistResult{}, fmt.Errorf("durable memo edits in GitOps mode require --output-path")
		}
		if input.AnnotateResource {
			return PersistResult{}, fmt.Errorf("resource annotation requires a direct cluster write and cannot be used with --output-path")
		}
		return s.writeManifestResult(obj, note.StoreType, input.OutputPath)
	}
	if err := s.kube.ApplyUnstructured(ctx, obj); err != nil {
		return PersistResult{}, permissionError(err, "update "+strings.ToLower(note.StoreType)+" memo")
	}
	updated, err := s.kube.Dynamic().Resource(gvr).Namespace(ns).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		return PersistResult{}, err
	}
	note = toNote(*updated, note.StoreType)
	if s.shouldSyncAnnotations(noteToTarget(note), input.AnnotateResource) {
		if err := s.syncTargetAnnotations(ctx, noteToTarget(note), input.RuntimeNamespace, input.AnnotateResource, note.ID); err != nil {
			return PersistResult{}, err
		}
	}
	return PersistResult{Note: note}, nil
}

func (s *Service) RemoveNote(ctx context.Context, id string, expiredRuntimeOnly bool, runtimeNamespace string, removeAnnotations bool) (map[string]any, error) {
	removed := []string{}
	if expiredRuntimeOnly {
		notes, err := s.ListNotes(ctx, true, runtimeNamespace, nil)
		if err != nil {
			return nil, err
		}
		for _, note := range notes {
			if note.StoreType == "Runtime" && note.ExpiresAt != nil && note.ExpiresAt.Before(time.Now().UTC()) {
				if _, _, gvr, ns, err := s.getNoteResource(ctx, note.ID, true, runtimeNamespace); err == nil {
					_ = s.kube.Delete(ctx, gvr, ns, note.ID)
					removed = append(removed, note.ID)
					if s.shouldSyncAnnotations(noteToTarget(note), removeAnnotations) {
						_ = s.syncTargetAnnotations(ctx, noteToTarget(note), runtimeNamespace, removeAnnotations, "")
					}
				}
			}
		}
		return map[string]any{"removed": removed}, nil
	}
	note, _, gvr, ns, err := s.getNoteResource(ctx, id, false, runtimeNamespace)
	if err != nil {
		return nil, err
	}
	if err := s.kube.Delete(ctx, gvr, ns, id); err != nil {
		return nil, permissionError(err, "delete memo")
	}
	if s.shouldSyncAnnotations(noteToTarget(note), removeAnnotations) {
		if err := s.syncTargetAnnotations(ctx, noteToTarget(note), runtimeNamespace, removeAnnotations, ""); err != nil {
			return nil, err
		}
	}
	return map[string]any{"removed": []string{note.ID}}, nil
}

func (s *Service) ClearRuntime(ctx context.Context, expiredOnly bool, runtimeNamespace string) (map[string]any, error) {
	if expiredOnly {
		return s.RemoveNote(ctx, "", true, runtimeNamespace, false)
	}
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	list, err := s.kube.List(ctx, runtimeGVR, runtimeNamespace)
	if err != nil {
		return nil, permissionError(err, "read runtime memos for cleanup")
	}
	removed := []string{}
	for _, item := range list.Items {
		if err := s.kube.Delete(ctx, runtimeGVR, runtimeNamespace, item.GetName()); err == nil {
			removed = append(removed, item.GetName())
		}
	}
	return map[string]any{"removed": removed}, nil
}

func (s *Service) GetActivity(ctx context.Context, kind, namespace, name, runtimeNamespace string) ([]model.Note, error) {
	notes, err := s.FindNotes(ctx, "", "activity", kind, namespace, name, true, true, runtimeNamespace)
	if err != nil {
		return nil, err
	}
	filtered := []model.Note{}
	for _, note := range notes {
		if note.StoreType == "Runtime" && strings.EqualFold(note.NoteType, "activity") {
			filtered = append(filtered, note)
		}
	}
	return filtered, nil
}

func (s *Service) Export(ctx context.Context, path string, includeRuntime bool, runtimeNamespace string) ([]string, error) {
	notes, err := s.ListNotes(ctx, includeRuntime, runtimeNamespace, nil)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}
	written := []string{}
	for _, note := range notes {
		target := filepath.Join(path, noteExportRelativePath(note))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		data, err := yaml.Marshal(note.RawResource)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return nil, err
		}
		written = append(written, target)
	}
	return written, nil
}

func (s *Service) Import(ctx context.Context, path string) ([]string, error) {
	applied := []string{}
	err := filepath.WalkDir(path, func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext := strings.ToLower(filepath.Ext(d.Name())); ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if err := s.kube.ApplyYAML(ctx, string(content), ""); err != nil {
			return err
		}
		applied = append(applied, fullPath)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(applied)
	return applied, nil
}

func (s *Service) SyncGitOps(ctx context.Context, path, direction string, includeRuntime bool, runtimeNamespace string) (map[string]any, error) {
	switch strings.ToLower(direction) {
	case "import":
		applied, err := s.Import(ctx, path)
		return map[string]any{"direction": "import", "applied": applied}, err
	default:
		written, err := s.Export(ctx, path, includeRuntime, runtimeNamespace)
		return map[string]any{"direction": "export", "written": written}, err
	}
}

func (s *Service) syncTargetAnnotations(ctx context.Context, target model.Target, runtimeNamespace string, explicit bool, preferredID string) error {
	target, err := normalizeAnnotationTarget(target, explicit)
	if err != nil {
		return err
	}
	if target.Mode == "" {
		return nil
	}
	namespaceScope := []string{}
	if target.Namespace != "" {
		namespaceScope = []string{target.Namespace}
	}
	notes, err := s.ListNotes(ctx, true, runtimeNamespace, namespaceScope)
	if err != nil {
		return err
	}
	matches := []model.Note{}
	for _, note := range notes {
		if targetMatches(note, target) {
			matches = append(matches, note)
		}
	}
	annotations := annotationStateForNotes(matches, preferredID, target)
	err = s.kube.UpdateTargetAnnotations(ctx, target.APIVersion, target.Kind, target.Namespace, target.Name, func(existing map[string]string) map[string]string {
		for _, key := range []string{annotationEnabled, annotationHasNote, annotationNoteCount, annotationRuntimeCount, annotationSummary, annotationView, annotationRuntimeEnabled, "notes.kubememo.io/note-ref"} {
			delete(existing, key)
		}
		for key, value := range annotations {
			existing[key] = value
		}
		return existing
	})
	return permissionError(err, "patch resource annotations")
}

func normalizeAnnotationTarget(target model.Target, explicit bool) (model.Target, error) {
	switch target.Mode {
	case "resource":
		if strings.TrimSpace(target.APIVersion) == "" || strings.TrimSpace(target.Kind) == "" || strings.TrimSpace(target.Name) == "" {
			return model.Target{}, fmt.Errorf("resource annotation target is incomplete")
		}
		return target, nil
	case "namespace":
		if strings.TrimSpace(target.Namespace) == "" {
			return model.Target{}, fmt.Errorf("namespace annotation target is incomplete")
		}
		return model.Target{Mode: "resource", APIVersion: "v1", Kind: "Namespace", Name: target.Namespace}, nil
	default:
		if !explicit {
			return model.Target{}, nil
		}
		return model.Target{}, fmt.Errorf("resource annotations are not supported for app targets")
	}
}

func noteExportRelativePath(note model.Note) string {
	fileName := note.ID + ".yaml"
	if note.StoreType == "Runtime" {
		scope := runtimeStorageNamespace(note)
		return filepath.Join("runtime", scope, fileName)
	}
	switch note.TargetMode {
	case "namespace":
		return filepath.Join("namespaces", firstNonEmptyString(note.Namespace, "cluster"), fileName)
	case "app":
		appDir := firstNonEmptyString(note.AppName, note.AppInstance, "unknown-app")
		return filepath.Join("apps", slugify(appDir), fileName)
	default:
		resourceDir := slugify(strings.ToLower(firstNonEmptyString(note.Kind, "resource")) + "-" + firstNonEmptyString(note.Name, note.ID))
		namespaceDir := firstNonEmptyString(note.Namespace, "cluster")
		return filepath.Join("resources", namespaceDir, resourceDir, fileName)
	}
}

func runtimeStorageNamespace(note model.Note) string {
	if metadata, ok := note.RawResource["metadata"].(map[string]any); ok {
		if ns, ok := metadata["namespace"].(string); ok && strings.TrimSpace(ns) != "" {
			return ns
		}
	}
	return firstNonEmptyString(note.Namespace, note.Name, "cluster")
}

func permissionError(err error, action string) error {
	if err == nil {
		return nil
	}
	if k8serrors.IsForbidden(err) {
		return fmt.Errorf("permission denied: cannot %s with current Kubernetes RBAC", action)
	}
	return err
}

func (s *Service) shouldSyncAnnotations(target model.Target, explicit bool) bool {
	if explicit {
		return true
	}
	return strings.EqualFold(target.Mode, "resource") || strings.EqualFold(target.Mode, "namespace")
}

func noteToTarget(note model.Note) model.Target {
	return model.Target{
		Mode:        note.TargetMode,
		APIVersion:  note.APIVersion,
		Kind:        note.Kind,
		Namespace:   note.Namespace,
		Name:        note.Name,
		AppName:     note.AppName,
		AppInstance: note.AppInstance,
	}
}

func targetMatches(note model.Note, target model.Target) bool {
	switch target.Mode {
	case "resource":
		return strings.EqualFold(note.TargetMode, "resource") &&
			strings.EqualFold(note.APIVersion, target.APIVersion) &&
			strings.EqualFold(note.Kind, target.Kind) &&
			note.Namespace == target.Namespace &&
			note.Name == target.Name
	default:
		return false
	}
}

func annotationStateForNotes(notes []model.Note, preferredID string, target model.Target) map[string]string {
	if len(notes) == 0 {
		return map[string]string{}
	}
	if preferredID != "" {
		sort.SliceStable(notes, func(i, j int) bool {
			if notes[i].ID == preferredID {
				return true
			}
			if notes[j].ID == preferredID {
				return false
			}
			return notes[i].ID < notes[j].ID
		})
	} else {
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].ID < notes[j].ID
		})
	}
	summaryNotes := append([]model.Note(nil), notes...)
	if preferredID != "" {
		sort.SliceStable(summaryNotes, func(i, j int) bool {
			if summaryNotes[i].ID == preferredID {
				return true
			}
			if summaryNotes[j].ID == preferredID {
				return false
			}
			return noteSortTime(summaryNotes[i]).After(noteSortTime(summaryNotes[j]))
		})
	} else {
		sort.Slice(summaryNotes, func(i, j int) bool {
			return noteSortTime(summaryNotes[i]).After(noteSortTime(summaryNotes[j]))
		})
	}
	ids := make([]string, 0, len(notes))
	hasRuntime := false
	runtimeCount := 0
	durableCount := 0
	durableSummary := ""
	runtimeSummary := ""
	for _, note := range summaryNotes {
		if note.StoreType == "Runtime" && runtimeSummary == "" {
			runtimeSummary = firstNonEmptyString(note.Summary, note.Title)
		}
		if note.StoreType == "Durable" && durableSummary == "" {
			durableSummary = firstNonEmptyString(note.Summary, note.Title)
		}
	}
	for _, note := range notes {
		ids = append(ids, note.ID)
		if note.StoreType == "Runtime" {
			hasRuntime = true
			runtimeCount++
		} else {
			durableCount++
		}
	}
	state := map[string]string{
		annotationEnabled:   "true",
		annotationHasNote:   "true",
		annotationNoteCount: fmt.Sprintf("%d", len(notes)),
		annotationView:      annotationViewCommand(target),
	}
	if runtimeCount > 0 {
		state[annotationRuntimeCount] = fmt.Sprintf("%d", runtimeCount)
		state[annotationRuntimeEnabled] = "true"
	}
	summary := ""
	switch {
	case len(notes) == 1:
		summary = firstNonEmptyString(durableSummary, runtimeSummary)
	case durableCount == 1 && runtimeCount == 0:
		summary = durableSummary
	}
	summary = truncateSummary(summary, maxAnnotationSummaryLen)
	if summary != "" {
		state[annotationSummary] = summary
	}
	_ = ids
	_ = hasRuntime
	return state
}

func annotationViewCommand(target model.Target) string {
	if strings.EqualFold(target.Kind, "Namespace") && strings.TrimSpace(target.Namespace) == "" {
		return fmt.Sprintf("kubememo show --kind Namespace --name %s", target.Name)
	}
	command := fmt.Sprintf("kubememo show --kind %s", target.Kind)
	if strings.TrimSpace(target.Namespace) != "" {
		command += " --namespace " + target.Namespace
	}
	if strings.TrimSpace(target.Name) != "" {
		command += " --name " + target.Name
	}
	return command
}

func noteSortTime(note model.Note) time.Time {
	if note.UpdatedAt != nil {
		return note.UpdatedAt.UTC()
	}
	if note.CreatedAt != nil {
		return note.CreatedAt.UTC()
	}
	return time.Time{}
}

func truncateSummary(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if maxLen <= 0 || len([]rune(text)) <= maxLen {
		return text
	}
	runes := []rune(text)
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return strings.TrimSpace(string(runes[:maxLen-3])) + "..."
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (s *Service) writeManifestResult(obj *unstructured.Unstructured, storeType, outputPath string) (PersistResult, error) {
	data, err := yaml.Marshal(obj.Object)
	if err != nil {
		return PersistResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return PersistResult{}, err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return PersistResult{}, err
	}
	note := toNote(*obj, storeType)
	return PersistResult{
		Note:       note,
		OutputPath: outputPath,
		Manifest:   string(data),
	}, nil
}

func (s *Service) getNoteResource(ctx context.Context, id string, runtime bool, runtimeNamespace string) (model.Note, *unstructured.Unstructured, schema.GroupVersionResource, string, error) {
	search := []struct {
		gvr   schema.GroupVersionResource
		ns    string
		store string
	}{
		{gvr: durableGVR, ns: "", store: "Durable"},
	}
	if runtime {
		search = []struct {
			gvr   schema.GroupVersionResource
			ns    string
			store string
		}{{gvr: runtimeGVR, ns: runtimeNamespace, store: "Runtime"}}
	} else {
		search = append(search, struct {
			gvr   schema.GroupVersionResource
			ns    string
			store string
		}{gvr: runtimeGVR, ns: runtimeNamespace, store: "Runtime"})
	}
	for _, candidate := range search {
		namespaces := []string{candidate.ns}
		if candidate.ns == "" {
			discovered, err := s.noteSearchNamespaces(ctx, candidate.gvr)
			if err != nil {
				return model.Note{}, nil, schema.GroupVersionResource{}, "", err
			}
			namespaces = discovered
		}
		for _, ns := range namespaces {
			resourceClient := s.kube.Dynamic().Resource(candidate.gvr)
			var obj *unstructured.Unstructured
			var err error
			if ns != "" {
				obj, err = resourceClient.Namespace(ns).Get(ctx, id, metav1.GetOptions{})
			} else {
				obj, err = resourceClient.Get(ctx, id, metav1.GetOptions{})
			}
			if err == nil {
				note := toNote(*obj, candidate.store)
				return note, obj, candidate.gvr, obj.GetNamespace(), nil
			}
		}
	}
	return model.Note{}, nil, schema.GroupVersionResource{}, "", fmt.Errorf("memo %q was not found", id)
}

func (s *Service) noteSearchNamespaces(ctx context.Context, gvr schema.GroupVersionResource) ([]string, error) {
	list, err := s.kube.List(ctx, gvr, "")
	if err != nil {
		if k8serrors.IsForbidden(err) {
			return []string{s.kube.CurrentNamespace()}, nil
		}
		return nil, err
	}
	seen := map[string]struct{}{}
	namespaces := []string{}
	for _, item := range list.Items {
		ns := item.GetNamespace()
		if ns == "" {
			continue
		}
		if _, ok := seen[ns]; ok {
			continue
		}
		seen[ns] = struct{}{}
		namespaces = append(namespaces, ns)
	}
	if len(namespaces) == 0 {
		namespaces = append(namespaces, s.kube.CurrentNamespace())
	}
	return namespaces, nil
}

func buildNoteObject(name string, input NewNoteInput, actor string, now time.Time, gitOpsEnabled bool) *unstructured.Unstructured {
	apiVersion := model.DurableAPIVersion
	kind := model.DurableKind
	namespace := input.Target.Namespace
	spec := map[string]any{
		"title":     input.Title,
		"summary":   input.Summary,
		"content":   input.Content,
		"format":    input.Format,
		"noteType":  input.NoteType,
		"severity":  input.Severity,
		"tags":      toAnySlice(input.Tags),
		"links":     []any{},
		"validFrom": now.Format(time.RFC3339),
		"createdBy": actor,
		"updatedBy": actor,
		"owner": map[string]any{
			"team":    input.OwnerTeam,
			"contact": input.OwnerContact,
		},
		"target": map[string]any{
			"mode":       input.Target.Mode,
			"apiVersion": input.Target.APIVersion,
			"kind":       input.Target.Kind,
			"namespace":  input.Target.Namespace,
			"name":       input.Target.Name,
			"appRef": map[string]any{
				"name":     input.Target.AppName,
				"instance": input.Target.AppInstance,
			},
		},
	}
	if input.ExpiresAt != nil {
		spec["expiresAt"] = input.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if input.Temporary {
		apiVersion = model.RuntimeAPIVersion
		kind = model.RuntimeKind
		namespace = input.RuntimeNamespace
		spec["temporary"] = true
		spec["createdAt"] = now.Format(time.RFC3339)
		spec["source"] = map[string]any{
			"type":       "manual",
			"generator":  "kubememo new",
			"confidence": "high",
		}
	} else {
		sourceType := "manual"
		if gitOpsEnabled {
			sourceType = "git"
		}
		spec["source"] = map[string]any{
			"type": sourceType,
		}
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
			"labels": map[string]any{
				"notes.kubememo.io/type": input.NoteType,
			},
		},
		"spec": spec,
		"status": map[string]any{
			"state":   "active",
			"expired": false,
		},
	}}
}

func toNote(obj unstructured.Unstructured, store string) model.Note {
	spec, _, _ := unstructured.NestedMap(obj.Object, "spec")
	owner := nestedMap(spec, "owner")
	target := nestedMap(spec, "target")
	appRef := nestedMap(target, "appRef")
	source := nestedMap(spec, "source")
	gitSource := nestedMap(source, "git")
	activity := nestedMap(spec, "activity")
	metadata := obj.Object["metadata"].(map[string]any)
	return model.Note{
		ID:              obj.GetName(),
		StoreType:       store,
		Title:           stringValue(spec["title"]),
		Summary:         stringValue(spec["summary"]),
		Content:         stringValue(spec["content"]),
		Format:          stringValue(spec["format"]),
		NoteType:        stringValue(spec["noteType"]),
		Temporary:       boolValue(spec["temporary"]),
		Severity:        stringValue(spec["severity"]),
		OwnerTeam:       stringValue(owner["team"]),
		OwnerContact:    stringValue(owner["contact"]),
		Tags:            stringSlice(spec["tags"]),
		TargetMode:      stringValue(target["mode"]),
		APIVersion:      stringValue(target["apiVersion"]),
		Kind:            stringValue(target["kind"]),
		Namespace:       stringValue(target["namespace"]),
		Name:            stringValue(target["name"]),
		AppName:         stringValue(appRef["name"]),
		AppInstance:     stringValue(appRef["instance"]),
		ValidFrom:       parseTime(spec["validFrom"]),
		ExpiresAt:       parseTime(spec["expiresAt"]),
		CreatedAt:       parseTime(firstNonEmpty(spec["createdAt"], metadata["creationTimestamp"])),
		UpdatedAt:       parseTime(metadata["creationTimestamp"]),
		CreatedBy:       stringValue(spec["createdBy"]),
		UpdatedBy:       stringValue(spec["updatedBy"]),
		SourceType:      stringValue(source["type"]),
		SourceGenerator: stringValue(source["generator"]),
		Confidence:      stringValue(source["confidence"]),
		GitRepo:         stringValue(gitSource["repo"]),
		GitPath:         stringValue(gitSource["path"]),
		GitRevision:     stringValue(gitSource["revision"]),
		Activity:        activity,
		RawResource:     obj.Object,
	}
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func firstNonEmpty(values ...any) any {
	for _, v := range values {
		if strings.TrimSpace(fmt.Sprintf("%v", v)) != "" && fmt.Sprintf("%v", v) != "<nil>" {
			return v
		}
	}
	return nil
}

func boolValue(v any) bool {
	value, ok := v.(bool)
	return ok && value
}

func stringSlice(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%v", item))
	}
	return out
}

func toAnySlice(items []string) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func parseTime(v any) *time.Time {
	if v == nil {
		return nil
	}
	text := strings.TrimSpace(fmt.Sprintf("%v", v))
	if text == "" || text == "<nil>" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return nil
	}
	return &parsed
}

func nestedMap(in map[string]any, key string) map[string]any {
	value, ok := in[key].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return value
}

func slugify(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", ".", "-", ":", "-", "'", "", "\"", "")
	text = replacer.Replace(text)
	builder := strings.Builder{}
	lastDash := false
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}
