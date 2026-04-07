package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/config"
	"github.com/KubeDeckio/KubeMemo/internal/kube"
	"github.com/KubeDeckio/KubeMemo/internal/model"
	appsv1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestClusterSmokeInstallCreateExportAndCleanup(t *testing.T) {
	if os.Getenv("KUBEMEMO_INTEGRATION") != "1" {
		t.Skip("set KUBEMEMO_INTEGRATION=1 to run cluster-backed integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg := config.Default()
	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	suffix := time.Now().UTC().Format("150405")
	targetNS := "kubememo-it-" + suffix
	runtimeNS := "kubememo-it-runtime-" + suffix

	t.Cleanup(func() {
		_ = svc.kube.Clientset().CoreV1().Namespaces().Delete(context.Background(), targetNS, metav1.DeleteOptions{})
		_ = svc.kube.Clientset().CoreV1().Namespaces().Delete(context.Background(), runtimeNS, metav1.DeleteOptions{})
	})

	if _, err := svc.Install(ctx, false, true, runtimeNS, false, false, false, ""); err != nil {
		t.Fatalf("install: %v", err)
	}

	if _, err := svc.kube.Clientset().CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: targetNS},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create target namespace: %v", err)
	}

	replicas := int32(1)
	_, err = svc.kube.Clientset().AppsV1().Deployments(targetNS).Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "orders-api", Namespace: targetNS},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "orders-api"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "orders-api"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "orders-api",
						Image: "nginx:stable",
						Ports: []corev1.ContainerPort{{ContainerPort: 80}},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(80)}},
							InitialDelaySeconds: 1,
							PeriodSeconds:       2,
						},
					}},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create deployment: %v", err)
	}

	durable, err := svc.NewNote(ctx, NewNoteInput{
		Title:            "Integration durable memo",
		Summary:          "Durable memo summary for integration coverage",
		Content:          "This durable memo validates install, annotate, get, and export behavior.",
		NoteType:         "advisory",
		Severity:         "info",
		RuntimeNamespace: runtimeNS,
		AnnotateResource: true,
		Target: model.Target{
			Mode:       "resource",
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  targetNS,
			Name:       "orders-api",
		},
	})
	if err != nil {
		t.Fatalf("create durable memo: %v", err)
	}

	runtimeExpiry := time.Now().UTC().Add(2 * time.Hour)
	runtime, err := svc.NewNote(ctx, NewNoteInput{
		Title:            "Integration runtime memo",
		Summary:          "Runtime memo summary for integration coverage",
		Content:          "This runtime memo validates runtime storage and combined retrieval.",
		NoteType:         "incident",
		Severity:         "warning",
		Temporary:        true,
		ExpiresAt:        &runtimeExpiry,
		RuntimeNamespace: runtimeNS,
		Target: model.Target{
			Mode:       "resource",
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  targetNS,
			Name:       "orders-api",
		},
	})
	if err != nil {
		t.Fatalf("create runtime memo: %v", err)
	}

	deployment, err := svc.kube.Clientset().AppsV1().Deployments(targetNS).Get(ctx, "orders-api", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if deployment.Annotations[annotationHasNote] != "true" {
		t.Fatalf("expected %s annotation to be true", annotationHasNote)
	}

	notes, err := svc.FindNotes(ctx, "", "", "Deployment", targetNS, "orders-api", true, false, runtimeNS)
	if err != nil {
		t.Fatalf("find notes: %v", err)
	}
	if len(notes) < 2 {
		t.Fatalf("expected both durable and runtime notes, got %d", len(notes))
	}

	exportDir := t.TempDir()
	written, err := svc.Export(ctx, exportDir, true, runtimeNS)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(written) < 2 {
		t.Fatalf("expected exported files, got %d", len(written))
	}
	expectedDurable := filepath.Join(exportDir, "resources", targetNS, "deployment-orders-api", durable.Note.ID+".yaml")
	if _, err := os.Stat(expectedDurable); err != nil {
		t.Fatalf("expected durable export path %s: %v", expectedDurable, err)
	}
	expectedRuntime := filepath.Join(exportDir, "runtime", runtimeNS, runtime.Note.ID+".yaml")
	if _, err := os.Stat(expectedRuntime); err != nil {
		t.Fatalf("expected runtime export path %s: %v", expectedRuntime, err)
	}

	if _, err := svc.RemoveNote(ctx, runtime.Note.ID, false, runtimeNS, false); err != nil {
		t.Fatalf("remove runtime memo: %v", err)
	}
	if _, err := svc.RemoveNote(ctx, durable.Note.ID, false, runtimeNS, true); err != nil {
		t.Fatalf("remove durable memo: %v", err)
	}
}

func TestRestrictedServiceAccountCanCreateMemoButCannotPatchAnnotations(t *testing.T) {
	if os.Getenv("KUBEMEMO_INTEGRATION") != "1" {
		t.Skip("set KUBEMEMO_INTEGRATION=1 to run cluster-backed integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg := config.Default()
	adminSvc, err := New(cfg)
	if err != nil {
		t.Fatalf("new admin service: %v", err)
	}

	suffix := time.Now().UTC().Format("150405")
	targetNS := "kubememo-rbac-" + suffix
	runtimeNS := "kubememo-rbac-runtime-" + suffix
	saName := "kubememo-limited"

	t.Cleanup(func() {
		_ = adminSvc.kube.Clientset().CoreV1().Namespaces().Delete(context.Background(), targetNS, metav1.DeleteOptions{})
		_ = adminSvc.kube.Clientset().CoreV1().Namespaces().Delete(context.Background(), runtimeNS, metav1.DeleteOptions{})
	})

	if _, err := adminSvc.Install(ctx, false, true, runtimeNS, false, false, false, ""); err != nil {
		t.Fatalf("install: %v", err)
	}

	if _, err := adminSvc.kube.Clientset().CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: targetNS},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create namespace %s: %v", targetNS, err)
	}

	replicas := int32(1)
	if _, err := adminSvc.kube.Clientset().AppsV1().Deployments(targetNS).Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "orders-api", Namespace: targetNS},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "orders-api"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "orders-api"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "orders-api", Image: "nginx:stable"}},
				},
			},
		},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create deployment: %v", err)
	}

	if _, err := adminSvc.kube.Clientset().CoreV1().ServiceAccounts(targetNS).Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: saName, Namespace: targetNS},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create service account: %v", err)
	}

	if _, err := adminSvc.kube.Clientset().RbacV1().Roles(targetNS).Create(ctx, &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "kubememo-limited", Namespace: targetNS},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"notes.kubememo.io"},
			Resources: []string{"memos"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
		}},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create role: %v", err)
	}

	if _, err := adminSvc.kube.Clientset().RbacV1().RoleBindings(targetNS).Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "kubememo-limited", Namespace: targetNS},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: saName, Namespace: targetNS}},
		RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: "kubememo-limited"},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create role binding: %v", err)
	}

	tokenResp, err := adminSvc.kube.Clientset().CoreV1().ServiceAccounts(targetNS).CreateToken(ctx, saName, &authv1.TokenRequest{}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	restCfg := adminSvc.kube.RestConfig()
	restCfg.BearerToken = tokenResp.Status.Token
	restCfg.BearerTokenFile = ""
	restCfg.CertFile = ""
	restCfg.KeyFile = ""
	restCfg.CAFile = ""
	restCfg.TLSClientConfig.CertFile = ""
	restCfg.TLSClientConfig.KeyFile = ""
	restCfg.TLSClientConfig.CertData = nil
	restCfg.TLSClientConfig.KeyData = nil
	restCfg.Username = ""
	restCfg.Password = ""
	restCfg.AuthProvider = nil
	restCfg.ExecProvider = nil
	limitedClient, err := kube.NewForConfig(restCfg, targetNS, "integration-rbac", adminSvc.kube.RawConfig())
	if err != nil {
		t.Fatalf("new limited client: %v", err)
	}
	limitedSvc := NewWithClient(cfg, limitedClient)

	status := limitedSvc.GetInstallationStatus(ctx, runtimeNS)
	if status.Status.Capabilities.DurableWrite.Allowed != true {
		t.Fatalf("expected durable write capability to be allowed, got %#v", status.Status.Capabilities.DurableWrite)
	}
	if status.Status.Capabilities.AnnotationPatch.Allowed {
		t.Fatalf("expected annotation patch to be denied for restricted service account")
	}

	_, err = limitedSvc.NewNote(ctx, NewNoteInput{
		Title:            "Limited durable memo",
		Summary:          "Created with restricted service account",
		Content:          "This memo should be creatable without annotation patch rights.",
		NoteType:         "advisory",
		Severity:         "info",
		RuntimeNamespace: runtimeNS,
		AnnotateResource: true,
		Target: model.Target{
			Mode:       "resource",
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  targetNS,
			Name:       "orders-api",
		},
	})
	if err == nil || err.Error() != "permission denied: cannot patch resource annotations with current Kubernetes RBAC" {
		t.Fatalf("expected annotation patch permission error, got %v", err)
	}
}
