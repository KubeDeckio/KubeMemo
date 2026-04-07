package kube

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	authv1 "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	restConfig     *rest.Config
	clientset      kubernetes.Interface
	dynamic        dynamic.Interface
	discovery      discovery.DiscoveryInterface
	mapper         *restmapper.DeferredDiscoveryRESTMapper
	rawConfig      api.Config
	currentNS      string
	currentContext string
}

func New() (*Client, error) {
	configLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, overrides)

	rawConfig, rawErr := clientConfig.RawConfig()
	namespace, _, nsErr := clientConfig.Namespace()
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			if rawErr != nil {
				return nil, rawErr
			}
			return nil, err
		}
		if strings.TrimSpace(namespace) == "" {
			namespace = os.Getenv("POD_NAMESPACE")
		}
	}
	if strings.TrimSpace(namespace) == "" || nsErr != nil {
		namespace = "default"
	}
	restConfig.Timeout = 20 * time.Second

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cached := memory.NewMemCacheClient(disco)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cached)

	return &Client{
		restConfig:     restConfig,
		clientset:      clientset,
		dynamic:        dyn,
		discovery:      disco,
		mapper:         mapper,
		rawConfig:      rawConfig,
		currentNS:      namespace,
		currentContext: rawConfig.CurrentContext,
	}, nil
}

func NewForConfig(restConfig *rest.Config, namespace, currentContext string, rawConfig api.Config) (*Client, error) {
	if strings.TrimSpace(namespace) == "" {
		namespace = "default"
	}
	restConfig = rest.CopyConfig(restConfig)
	restConfig.Timeout = 20 * time.Second

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cached := memory.NewMemCacheClient(disco)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cached)

	return &Client{
		restConfig:     restConfig,
		clientset:      clientset,
		dynamic:        dyn,
		discovery:      disco,
		mapper:         mapper,
		rawConfig:      rawConfig,
		currentNS:      namespace,
		currentContext: rawConfig.CurrentContext,
	}, nil
}

func (c *Client) Config() *rest.Config            { return c.restConfig }
func (c *Client) Dynamic() dynamic.Interface      { return c.dynamic }
func (c *Client) Clientset() kubernetes.Interface { return c.clientset }
func (c *Client) CurrentNamespace() string        { return c.currentNS }
func (c *Client) CurrentContext() string          { return c.currentContext }
func (c *Client) RestConfig() *rest.Config        { return rest.CopyConfig(c.restConfig) }
func (c *Client) RawConfig() api.Config           { return c.rawConfig }

func (c *Client) ServerVersion(ctx context.Context) (string, error) {
	info, err := c.discovery.ServerVersion()
	if err != nil {
		return "", err
	}
	return info.GitVersion, nil
}

func (c *Client) GetActor(ctx context.Context) string {
	review, err := c.clientset.AuthenticationV1().SelfSubjectReviews().Create(ctx, &authv1.SelfSubjectReview{}, metav1.CreateOptions{})
	if err == nil && strings.TrimSpace(review.Status.UserInfo.Username) != "" {
		return review.Status.UserInfo.Username
	}

	if ctxName := c.currentContext; strings.TrimSpace(ctxName) != "" {
		if ctxConfig, ok := c.rawConfig.Contexts[ctxName]; ok {
			if strings.TrimSpace(ctxConfig.AuthInfo) != "" {
				return ctxConfig.AuthInfo
			}
		}
	}

	for _, candidate := range []string{os.Getenv("KUBEMEMO_USER"), os.Getenv("USER"), os.Getenv("USERNAME")} {
		if strings.TrimSpace(candidate) != "" {
			return strings.TrimSpace(candidate)
		}
	}

	return "unknown"
}

func (c *Client) CanI(ctx context.Context, verb string, gvr schema.GroupVersionResource, namespace string) (bool, string) {
	review := &authzv1.SelfSubjectAccessReview{
		Spec: authzv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authzv1.ResourceAttributes{
				Verb:      verb,
				Group:     gvr.Group,
				Version:   gvr.Version,
				Resource:  gvr.Resource,
				Namespace: namespace,
			},
		},
	}
	result, err := c.clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, err.Error()
	}
	if result.Status.Allowed {
		return true, ""
	}
	if strings.TrimSpace(result.Status.Reason) != "" {
		return false, result.Status.Reason
	}
	return false, "access denied by Kubernetes RBAC"
}

func (c *Client) ResolveResource(apiVersion, kind string) (*meta.RESTMapping, schema.GroupVersionResource, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, schema.GroupVersionResource{}, err
	}
	mapping, err := c.mapper.RESTMapping(schema.GroupKind{Group: gv.Group, Kind: kind}, gv.Version)
	if err != nil {
		return nil, schema.GroupVersionResource{}, err
	}
	return mapping, mapping.Resource, nil
}

func (c *Client) GetTargetResource(ctx context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, *meta.RESTMapping, error) {
	mapping, _, err := c.ResolveResource(apiVersion, kind)
	if err != nil {
		return nil, nil, err
	}
	resourceClient := c.dynamic.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := namespace
		if strings.TrimSpace(ns) == "" {
			ns = c.currentNS
		}
		obj, err := resourceClient.Namespace(ns).Get(ctx, name, metav1.GetOptions{})
		return obj, mapping, err
	}
	obj, err := resourceClient.Get(ctx, name, metav1.GetOptions{})
	return obj, mapping, err
}

func (c *Client) UpdateTargetAnnotations(ctx context.Context, apiVersion, kind, namespace, name string, mutate func(map[string]string) map[string]string) error {
	obj, mapping, err := c.GetTargetResource(ctx, apiVersion, kind, namespace, name)
	if err != nil {
		return err
	}
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	updatedAnnotations := mutate(annotations)
	patchObj := map[string]any{
		"metadata": map[string]any{
			"annotations": updatedAnnotations,
		},
	}
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return err
	}
	resourceClient := c.dynamic.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := obj.GetNamespace()
		if strings.TrimSpace(ns) == "" {
			ns = namespace
		}
		if strings.TrimSpace(ns) == "" {
			ns = c.currentNS
		}
		_, err = resourceClient.Namespace(ns).Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
		return err
	}
	_, err = resourceClient.Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

func (c *Client) ApplyYAML(ctx context.Context, manifest string, namespaceOverride string) error {
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)
	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if len(obj.Object) == 0 {
			continue
		}
		if namespaceOverride != "" && obj.GetKind() == "Namespace" {
			obj.SetName(namespaceOverride)
		}
		if namespaceOverride != "" && obj.GetNamespace() == "kubememo-runtime" {
			obj.SetNamespace(namespaceOverride)
		}
		if err := c.ApplyUnstructured(ctx, &obj); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ApplyUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
	mapping, err := c.mapper.RESTMapping(obj.GroupVersionKind().GroupKind(), obj.GroupVersionKind().Version)
	if err != nil {
		return err
	}

	resourceClient := c.dynamic.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := obj.GetNamespace()
		if ns == "" {
			ns = c.currentNS
			obj.SetNamespace(ns)
		}
		_, err := resourceClient.Namespace(ns).Get(ctx, obj.GetName(), metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				_, err = resourceClient.Namespace(ns).Create(ctx, obj, metav1.CreateOptions{})
				return err
			}
			return err
		}
		existing, err := resourceClient.Namespace(ns).Get(ctx, obj.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = resourceClient.Namespace(ns).Update(ctx, obj, metav1.UpdateOptions{})
		return err
	}

	existing, err := resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			_, err = resourceClient.Create(ctx, obj, metav1.CreateOptions{})
			return err
		}
		return err
	}

	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
	return err
}

func (c *Client) Delete(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	resourceClient := c.dynamic.Resource(gvr)
	if namespace != "" {
		if err := resourceClient.Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		return nil
	}
	if err := resourceClient.Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (c *Client) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	resourceClient := c.dynamic.Resource(gvr)
	if namespace != "" {
		return resourceClient.Namespace(namespace).List(ctx, metav1.ListOptions{})
	}
	return resourceClient.List(ctx, metav1.ListOptions{})
}

func (c *Client) Watch(ctx context.Context, gvr schema.GroupVersionResource, namespace, resourceVersion string) (watch.Interface, error) {
	resourceClient := c.dynamic.Resource(gvr)
	opts := metav1.ListOptions{ResourceVersion: resourceVersion, Watch: true}
	if namespace != "" {
		return resourceClient.Namespace(namespace).Watch(ctx, opts)
	}
	return resourceClient.Watch(ctx, opts)
}

func KubeconfigPath() string {
	if env := os.Getenv("KUBECONFIG"); strings.TrimSpace(env) != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
