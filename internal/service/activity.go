package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

type ActivityEvent struct {
	Target       model.Target `json:"target"`
	Action       string       `json:"action"`
	FieldPath    string       `json:"fieldPath"`
	OldValue     string       `json:"oldValue"`
	NewValue     string       `json:"newValue"`
	Actor        string       `json:"actor,omitempty"`
	ActorType    string       `json:"actorType,omitempty"`
	Confidence   string       `json:"confidence,omitempty"`
	DetectedAt   time.Time    `json:"detectedAt"`
	NoteID       string       `json:"noteId,omitempty"`
	Deduplicated bool         `json:"deduplicated,omitempty"`
}

type watchTarget struct {
	kind       string
	apiVersion string
	gvr        schema.GroupVersionResource
}

func (s *Service) StartActivityCapture(ctx context.Context, runtimeNamespace string, namespaces, kinds []string, onEvent func(ActivityEvent)) error {
	if runtimeNamespace == "" {
		runtimeNamespace = s.cfg.RuntimeNamespace
	}
	targets := s.activityWatchTargets(kinds)
	if len(targets) == 0 {
		return fmt.Errorf("no watch kinds resolved for activity capture")
	}
	if len(namespaces) == 0 {
		namespaces = []string{""}
	}

	errCh := make(chan error, len(targets)*len(namespaces))
	for _, target := range targets {
		for _, ns := range namespaces {
			go func(target watchTarget, ns string) {
				errCh <- s.watchTargetLoop(ctx, target, ns, runtimeNamespace, onEvent)
			}(target, ns)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			if err == nil || ctx.Err() != nil {
				continue
			}
			return err
		}
	}
}

func (s *Service) watchTargetLoop(ctx context.Context, target watchTarget, namespace, runtimeNamespace string, onEvent func(ActivityEvent)) error {
	resourceVersion := ""
	state := map[string]*unstructured.Unstructured{}

	for {
		list, err := s.kube.List(ctx, target.gvr, namespace)
		if err != nil {
			return err
		}
		resourceVersion = list.GetResourceVersion()
		state = map[string]*unstructured.Unstructured{}
		for i := range list.Items {
			item := list.Items[i]
			key := objectKey(item)
			copy := item.DeepCopy()
			state[key] = copy
		}

		watcher, err := s.kube.Watch(ctx, target.gvr, namespace, resourceVersion)
		if err != nil {
			return err
		}

		restart := false
		for !restart {
			select {
			case <-ctx.Done():
				watcher.Stop()
				return nil
			case evt, ok := <-watcher.ResultChan():
				if !ok {
					restart = true
					break
				}

				obj, ok := evt.Object.(*unstructured.Unstructured)
				if !ok {
					continue
				}
				key := objectKey(*obj)
				switch evt.Type {
				case watch.Added:
					state[key] = obj.DeepCopy()
				case watch.Modified:
					oldObj := state[key]
					newObj := obj.DeepCopy()
					state[key] = newObj
					if oldObj == nil {
						continue
					}
					events, err := s.detectActivity(ctx, target, *oldObj, *newObj, runtimeNamespace)
					if err != nil {
						return err
					}
					for _, activity := range events {
						if onEvent != nil {
							onEvent(activity)
						}
					}
				case watch.Deleted:
					delete(state, key)
				case watch.Error:
					restart = true
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
}

func (s *Service) detectActivity(ctx context.Context, target watchTarget, oldObj, newObj unstructured.Unstructured, runtimeNamespace string) ([]ActivityEvent, error) {
	noteTarget := model.Target{
		Mode:       "resource",
		APIVersion: target.apiVersion,
		Kind:       target.kind,
		Namespace:  newObj.GetNamespace(),
		Name:       newObj.GetName(),
	}

	if s.cfg.ActivityCapture.RequireNotesEnabled {
		enabled, err := s.isTargetNoteEnabled(ctx, noteTarget, newObj, runtimeNamespace)
		if err != nil {
			return nil, err
		}
		if !enabled {
			return nil, nil
		}
	}

	events := activityDiffs(target.kind, target.apiVersion, oldObj.Object, newObj.Object)
	if len(events) == 0 {
		return nil, nil
	}

	captured := []ActivityEvent{}
	for _, evt := range events {
		evt.Target = noteTarget
		evt.DetectedAt = time.Now().UTC()
		evt.Actor, evt.ActorType = inferActor(newObj)
		evt.Confidence = "medium"

		duplicate, err := s.isDuplicateActivity(ctx, evt, runtimeNamespace)
		if err != nil {
			return nil, err
		}
		if duplicate {
			evt.Deduplicated = true
			captured = append(captured, evt)
			continue
		}

		note, err := s.createActivityMemo(ctx, evt, runtimeNamespace)
		if err != nil {
			return nil, err
		}
		evt.NoteID = note.ID
		captured = append(captured, evt)
	}
	return captured, nil
}

func (s *Service) createActivityMemo(ctx context.Context, evt ActivityEvent, runtimeNamespace string) (model.Note, error) {
	title := activityTitle(evt)
	summary := activitySummary(evt)
	content := "Automatically captured operational activity."
	if evt.OldValue != "" || evt.NewValue != "" {
		content = fmt.Sprintf("Detected a %s change on %s.\n\nOld: %s\nNew: %s", evt.Action, renderTarget(evt.Target), evt.OldValue, evt.NewValue)
	}
	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.Runtime.DefaultExpiryHours) * time.Hour)
	input := NewNoteInput{
		Title:            title,
		Summary:          summary,
		Content:          content,
		Format:           "markdown",
		NoteType:         "activity",
		Severity:         "info",
		Temporary:        true,
		RuntimeNamespace: runtimeNamespace,
		ExpiresAt:        &expiresAt,
		Target:           evt.Target,
	}
	resourceName := slugify(fmt.Sprintf("%s-%s-%d", evt.Target.Name, evt.Action, evt.DetectedAt.UnixNano()))
	actor := s.GetActor(ctx)
	obj := buildNoteObject(resourceName, input, actor, evt.DetectedAt, false)
	spec := nestedMap(obj.Object, "spec")
	spec["source"] = map[string]any{
		"type":       "auto",
		"generator":  "activity-capture",
		"confidence": evt.Confidence,
	}
	spec["activity"] = map[string]any{
		"action":     evt.Action,
		"fieldPath":  evt.FieldPath,
		"oldValue":   evt.OldValue,
		"newValue":   evt.NewValue,
		"actor":      evt.Actor,
		"actorType":  evt.ActorType,
		"detectedAt": evt.DetectedAt.Format(time.RFC3339),
	}
	_ = unstructured.SetNestedMap(obj.Object, spec, "spec")
	obj.SetNamespace(runtimeNamespace)
	if err := s.kube.ApplyUnstructured(ctx, obj); err != nil {
		return model.Note{}, err
	}
	created, err := s.kube.Dynamic().Resource(runtimeGVR).Namespace(runtimeNamespace).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		return model.Note{}, err
	}
	return toNote(*created, "Runtime"), nil
}

func (s *Service) isDuplicateActivity(ctx context.Context, evt ActivityEvent, runtimeNamespace string) (bool, error) {
	notes, err := s.GetActivity(ctx, evt.Target.Kind, evt.Target.Namespace, evt.Target.Name, runtimeNamespace)
	if err != nil {
		return false, err
	}
	window := time.Duration(s.cfg.ActivityCapture.DedupeWindowSeconds) * time.Second
	for _, note := range notes {
		activity := note.Activity
		if !strings.EqualFold(stringValue(activity["action"]), evt.Action) {
			continue
		}
		if stringValue(activity["fieldPath"]) != evt.FieldPath {
			continue
		}
		if stringValue(activity["oldValue"]) != evt.OldValue || stringValue(activity["newValue"]) != evt.NewValue {
			continue
		}
		if note.CreatedAt != nil && evt.DetectedAt.Sub(note.CreatedAt.UTC()) <= window {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) isTargetNoteEnabled(ctx context.Context, target model.Target, obj unstructured.Unstructured, runtimeNamespace string) (bool, error) {
	annotations := obj.GetAnnotations()
	for _, key := range []string{annotationEnabled, annotationHasNote, annotationRuntimeEnabled} {
		if strings.EqualFold(annotations[key], "true") {
			return true, nil
		}
	}
	notes, err := s.FindNotes(ctx, "", "", target.Kind, target.Namespace, target.Name, true, true, runtimeNamespace)
	if err != nil {
		return false, err
	}
	return len(notes) > 0, nil
}

func (s *Service) activityWatchTargets(kinds []string) []watchTarget {
	requested := map[string]struct{}{}
	for _, kind := range kinds {
		if strings.TrimSpace(kind) != "" {
			requested[strings.ToLower(strings.TrimSpace(kind))] = struct{}{}
		}
	}
	targets := []watchTarget{}
	seen := map[string]struct{}{}
	for _, kind := range s.cfg.ActivityCapture.WatchKinds {
		if len(requested) > 0 {
			if _, ok := requested[strings.ToLower(kind)]; !ok {
				continue
			}
		}
		apiVersion := inferTargetAPIVersion(kind, "")
		_, gvr, err := s.kube.ResolveResource(apiVersion, kind)
		if err != nil {
			continue
		}
		key := gvr.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, watchTarget{
			kind:       kind,
			apiVersion: apiVersion,
			gvr:        gvr,
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].kind < targets[j].kind
	})
	return targets
}

func activityDiffs(kind, apiVersion string, oldObj, newObj map[string]any) []ActivityEvent {
	var candidates []ActivityEvent
	switch strings.ToLower(kind) {
	case "deployment", "statefulset", "daemonset", "replicaset", "horizontalpodautoscaler":
		candidates = append([]ActivityEvent{},
			replicaDiff(oldObj, newObj),
			imageDiff(oldObj, newObj),
			resourceDiff(oldObj, newObj),
			nodeSelectorDiff(oldObj, newObj),
			tolerationsDiff(oldObj, newObj),
		)
	case "service":
		candidates = append([]ActivityEvent{}, serviceTypeDiff(oldObj, newObj))
	case "ingress":
		candidates = append([]ActivityEvent{}, ingressDiff(oldObj, newObj))
	default:
		return nil
	}
	filtered := make([]ActivityEvent, 0, len(candidates))
	for _, evt := range candidates {
		if strings.TrimSpace(evt.Action) == "" {
			continue
		}
		filtered = append(filtered, evt)
	}
	return filtered
}

func replicaDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal, oldOK := nestedInt(oldObj, "spec", "replicas")
	newVal, newOK := nestedInt(newObj, "spec", "replicas")
	if !oldOK && !newOK {
		return ActivityEvent{}
	}
	if oldVal == newVal {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "scale", FieldPath: "spec.replicas", OldValue: strconv.FormatInt(oldVal, 10), NewValue: strconv.FormatInt(newVal, 10)}
}

func imageDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal := containerImageSummary(oldObj)
	newVal := containerImageSummary(newObj)
	if oldVal == newVal || oldVal == "" && newVal == "" {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "imageChange", FieldPath: "spec.template.spec.containers[].image", OldValue: oldVal, NewValue: newVal}
}

func resourceDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal := containerResourceSummary(oldObj)
	newVal := containerResourceSummary(newObj)
	if oldVal == newVal || oldVal == "" && newVal == "" {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "resourceChange", FieldPath: "spec.template.spec.containers[].resources", OldValue: oldVal, NewValue: newVal}
}

func serviceTypeDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal, _, _ := unstructured.NestedString(oldObj, "spec", "type")
	newVal, _, _ := unstructured.NestedString(newObj, "spec", "type")
	if oldVal == newVal {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "serviceTypeChange", FieldPath: "spec.type", OldValue: oldVal, NewValue: newVal}
}

func ingressDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal := ingressSummary(oldObj)
	newVal := ingressSummary(newObj)
	if oldVal == newVal || oldVal == "" && newVal == "" {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "ingressChange", FieldPath: "spec.rules", OldValue: oldVal, NewValue: newVal}
}

func nodeSelectorDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal := mapSummary(oldObj, "spec", "template", "spec", "nodeSelector")
	newVal := mapSummary(newObj, "spec", "template", "spec", "nodeSelector")
	if oldVal == newVal || oldVal == "" && newVal == "" {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "nodeSelectorChange", FieldPath: "spec.template.spec.nodeSelector", OldValue: oldVal, NewValue: newVal}
}

func tolerationsDiff(oldObj, newObj map[string]any) ActivityEvent {
	oldVal := sliceSummary(oldObj, "spec", "template", "spec", "tolerations")
	newVal := sliceSummary(newObj, "spec", "template", "spec", "tolerations")
	if oldVal == newVal || oldVal == "" && newVal == "" {
		return ActivityEvent{}
	}
	return ActivityEvent{Action: "tolerationChange", FieldPath: "spec.template.spec.tolerations", OldValue: oldVal, NewValue: newVal}
}

func nestedInt(obj map[string]any, fields ...string) (int64, bool) {
	val, found, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil || !found || val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

func containerImageSummary(obj map[string]any) string {
	items := []string{}
	for _, path := range [][]string{{"spec", "template", "spec", "containers"}, {"spec", "template", "spec", "initContainers"}} {
		containers, found, _ := unstructured.NestedSlice(obj, path...)
		if !found {
			continue
		}
		for _, item := range containers {
			container, ok := item.(map[string]any)
			if !ok {
				continue
			}
			items = append(items, fmt.Sprintf("%s=%s", stringValue(container["name"]), stringValue(container["image"])))
		}
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func containerResourceSummary(obj map[string]any) string {
	containers, found, _ := unstructured.NestedSlice(obj, "spec", "template", "spec", "containers")
	if !found {
		return ""
	}
	items := []string{}
	for _, item := range containers {
		container, ok := item.(map[string]any)
		if !ok {
			continue
		}
		resources, _ := container["resources"].(map[string]any)
		if len(resources) == 0 {
			continue
		}
		data, err := json.Marshal(resources)
		if err != nil {
			continue
		}
		items = append(items, fmt.Sprintf("%s=%s", stringValue(container["name"]), string(data)))
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func ingressSummary(obj map[string]any) string {
	rules, found, _ := unstructured.NestedSlice(obj, "spec", "rules")
	if !found {
		return ""
	}
	items := []string{}
	for _, item := range rules {
		rule, ok := item.(map[string]any)
		if !ok {
			continue
		}
		host := stringValue(rule["host"])
		httpMap, _ := rule["http"].(map[string]any)
		pathsAny, _ := httpMap["paths"].([]any)
		paths := []string{}
		for _, p := range pathsAny {
			pathMap, ok := p.(map[string]any)
			if !ok {
				continue
			}
			paths = append(paths, stringValue(pathMap["path"]))
		}
		sort.Strings(paths)
		items = append(items, fmt.Sprintf("%s:%s", host, strings.Join(paths, "|")))
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func mapSummary(obj map[string]any, fields ...string) string {
	val, found, _ := unstructured.NestedStringMap(obj, fields...)
	if !found || len(val) == 0 {
		return ""
	}
	keys := make([]string, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, val[k]))
	}
	return strings.Join(parts, ", ")
}

func sliceSummary(obj map[string]any, fields ...string) string {
	val, found, _ := unstructured.NestedSlice(obj, fields...)
	if !found || len(val) == 0 {
		return ""
	}
	data, err := json.Marshal(val)
	if err != nil {
		return ""
	}
	return string(data)
}

func inferActor(obj unstructured.Unstructured) (string, string) {
	managedFields := obj.GetManagedFields()
	if len(managedFields) > 0 {
		latest := managedFields[0]
		for _, entry := range managedFields[1:] {
			if entry.Time != nil && (latest.Time == nil || entry.Time.After(latest.Time.Time)) {
				latest = entry
			}
		}
		manager := strings.TrimSpace(latest.Manager)
		if manager != "" {
			actorType := "user"
			if strings.Contains(manager, "controller") || strings.Contains(manager, "deployment") || strings.Contains(manager, "replicaset") {
				actorType = "controller"
			}
			return manager, actorType
		}
	}
	return "unknown", "unknown"
}

func activityTitle(evt ActivityEvent) string {
	switch evt.Action {
	case "scale":
		return "Manual scale change detected"
	case "imageChange":
		return "Container image change detected"
	case "serviceTypeChange":
		return "Service type change detected"
	case "ingressChange":
		return "Ingress rule change detected"
	default:
		return "Operational activity detected"
	}
}

func activitySummary(evt ActivityEvent) string {
	if evt.OldValue != "" || evt.NewValue != "" {
		return fmt.Sprintf("%s changed from %s to %s", evt.FieldPath, evt.OldValue, evt.NewValue)
	}
	return fmt.Sprintf("%s change detected", evt.Action)
}

func renderTarget(target model.Target) string {
	label := target.Kind
	if target.Namespace != "" {
		label += "/" + target.Namespace
	}
	if target.Name != "" {
		label += "/" + target.Name
	}
	return label
}

func objectKey(obj unstructured.Unstructured) string {
	return fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
}
