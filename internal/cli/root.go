package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/config"
	"github.com/KubeDeckio/KubeMemo/internal/model"
	"github.com/KubeDeckio/KubeMemo/internal/render"
	"github.com/KubeDeckio/KubeMemo/internal/service"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	cfg     config.Config
	service *service.Service
}

var version = "dev"

func NewRootCommand() *cobra.Command {
	opts := &rootOptions{cfg: config.Default()}
	cmd := &cobra.Command{
		Use:   "kubememo",
		Short: "KubeMemo is a Kubernetes operational memory tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.service != nil {
				return nil
			}
			svc, err := service.New(opts.cfg)
			if err != nil {
				return err
			}
			opts.service = svc
			return nil
		},
	}

	cmd.AddCommand(
		newInstallCmd(opts),
		newUninstallCmd(opts),
		newUpdateCmd(opts),
		newTestInstallationCmd(opts),
		newStatusCmd(opts),
		newVersionCmd(),
		newGetCmd(opts),
		newFindCmd(opts),
		newShowCmd(opts),
		newNewCmd(opts),
		newSetCmd(opts),
		newRemoveCmd(opts),
		newExportCmd(opts),
		newImportCmd(opts),
		newSyncCmd(opts),
		newTestGitOpsCmd(opts),
		newTestRuntimeStoreCmd(opts),
		newClearCmd(opts),
		newActivityCmd(opts),
		newWatchActivityCmd(opts),
		newConfigCmd(opts),
		newTuiCmd(opts),
	)

	return cmd
}

func newVersionCmd() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the KubeMemo version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeOutput(output, map[string]any{
				"version": version,
			})
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newInstallCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace, activityCaptureImage string
	var durableOnly, enableRuntimeStore, installRbac, gitOpsAware, enableActivityCapture bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install KubeMemo cluster prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := opts.service.Install(context.Background(), durableOnly, enableRuntimeStore, runtimeNamespace, installRbac, gitOpsAware, enableActivityCapture, activityCaptureImage)
			if err != nil {
				return err
			}
			return writeOutput(output, status)
		},
	}
	cmd.Flags().BoolVar(&durableOnly, "durable-only", false, "Install only the durable store")
	cmd.Flags().BoolVar(&enableRuntimeStore, "enable-runtime-store", false, "Install the runtime store")
	cmd.Flags().BoolVar(&installRbac, "install-rbac", false, "Install bundled RBAC")
	cmd.Flags().BoolVar(&gitOpsAware, "gitops-aware", false, "Adjust install behavior for GitOps clusters")
	cmd.Flags().BoolVar(&enableActivityCapture, "enable-activity-capture", false, "Install the optional in-cluster activity capture deployment")
	cmd.Flags().StringVar(&activityCaptureImage, "activity-capture-image", opts.cfg.ActivityCapture.Image, "Container image for in-cluster activity capture")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newUninstallCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace string
	var runtimeOnly, removeData bool
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall KubeMemo prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.service.Uninstall(context.Background(), runtimeOnly, removeData, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeOutput(output, result)
		},
	}
	cmd.Flags().BoolVar(&runtimeOnly, "runtime-only", false, "Remove runtime components only")
	cmd.Flags().BoolVar(&removeData, "remove-data", false, "Delete memo objects before uninstalling")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newUpdateCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace, activityCaptureImage string
	var includeRbac, gitOpsAware, enableActivityCapture bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update installed KubeMemo prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := opts.service.Update(context.Background(), includeRbac, runtimeNamespace, gitOpsAware, enableActivityCapture, activityCaptureImage)
			if err != nil {
				return err
			}
			return writeOutput(output, status)
		},
	}
	cmd.Flags().BoolVar(&includeRbac, "include-rbac", false, "Update RBAC resources as well")
	cmd.Flags().BoolVar(&gitOpsAware, "gitops-aware", false, "Adjust update behavior for GitOps clusters")
	cmd.Flags().BoolVar(&enableActivityCapture, "enable-activity-capture", false, "Install or update the optional in-cluster activity capture deployment")
	cmd.Flags().StringVar(&activityCaptureImage, "activity-capture-image", opts.cfg.ActivityCapture.Image, "Container image for in-cluster activity capture")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newTestInstallationCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace string
	cmd := &cobra.Command{
		Use:   "test-installation",
		Short: "Test KubeMemo installation state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeOutput(output, opts.service.TestInstallation(context.Background(), runtimeNamespace))
		},
	}
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newStatusCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get KubeMemo installation status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeOutput(output, opts.service.GetInstallationStatus(context.Background(), runtimeNamespace))
		},
	}
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newGetCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace string
	var includeRuntime bool
	var namespaces []string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get memos",
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := opts.service.ListNotes(context.Background(), includeRuntime, runtimeNamespace, namespaces)
			if err != nil {
				return err
			}
			return writeOutput(output, model.NoteList{Items: notes})
		},
	}
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", false, "Include runtime memos")
	cmd.Flags().StringSliceVar(&namespaces, "namespace", nil, "Namespace scope")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newFindCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace, text, noteType, kind, namespace, name string
	includeRuntime := true
	var includeExpired bool
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find memos by filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := opts.service.FindNotes(context.Background(), text, noteType, kind, namespace, name, includeRuntime, includeExpired, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeOutput(output, model.NoteList{Items: notes})
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "Text filter")
	cmd.Flags().StringVar(&noteType, "note-type", "", "Note type")
	cmd.Flags().StringVar(&kind, "kind", "", "Target kind")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Target namespace")
	cmd.Flags().StringVar(&name, "name", "", "Target name")
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", true, "Include runtime memos")
	cmd.Flags().BoolVar(&includeExpired, "include-expired", false, "Include expired runtime memos")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newShowCmd(opts *rootOptions) *cobra.Command {
	var runtimeNamespace, kind, namespace, name string
	includeRuntime := true
	var noColor bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Render memos as terminal cards",
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := opts.service.FindNotes(context.Background(), "", "", kind, namespace, name, includeRuntime, false, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeMaybePaged(render.RenderNotes(notes, render.CardOptions{Header: true, NoColor: noColor, Width: 78}))
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Target kind")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Target namespace")
	cmd.Flags().StringVar(&name, "name", "", "Target name")
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", true, "Include runtime memos")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable ANSI color")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	return cmd
}

func newNewCmd(opts *rootOptions) *cobra.Command {
	var output string
	var input service.NewNoteInput
	var kind, namespace, name, apiVersion, targetNamespace, appName, appInstance, outputPath string
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a memo",
		RunE: func(cmd *cobra.Command, args []string) error {
			input.Target = opts.service.ResolveTarget(kind, namespace, name, apiVersion, targetNamespace, appName, appInstance)
			input.OutputPath = outputPath
			result, err := opts.service.NewNote(context.Background(), input)
			if err != nil {
				return err
			}
			return writeOutput(output, result)
		},
	}
	cmd.Flags().StringVar(&input.Title, "title", "", "Title")
	cmd.Flags().StringVar(&input.Summary, "summary", "", "Summary")
	cmd.Flags().StringVar(&input.Content, "content", "", "Content")
	cmd.Flags().StringVar(&input.Format, "format", "markdown", "Format")
	cmd.Flags().StringVar(&input.NoteType, "note-type", "advisory", "Note type")
	cmd.Flags().StringVar(&input.Severity, "severity", "info", "Severity")
	cmd.Flags().StringVar(&input.OwnerTeam, "owner-team", "", "Owner team")
	cmd.Flags().StringVar(&input.OwnerContact, "owner-contact", "", "Owner contact")
	cmd.Flags().StringSliceVar(&input.Tags, "tag", nil, "Tags")
	cmd.Flags().BoolVar(&input.Temporary, "temporary", false, "Create a runtime memo")
	cmd.Flags().StringVar(&input.RuntimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	cmd.Flags().StringVar(&kind, "kind", "", "Target kind")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Target namespace")
	cmd.Flags().StringVar(&name, "name", "", "Target name")
	cmd.Flags().StringVar(&apiVersion, "api-version", "v1", "Target apiVersion")
	cmd.Flags().StringVar(&targetNamespace, "target-namespace", "", "Namespace target")
	cmd.Flags().StringVar(&appName, "app-name", "", "App target name")
	cmd.Flags().StringVar(&appInstance, "app-instance", "", "App target instance")
	cmd.Flags().StringVar(&outputPath, "output-path", "", "Write durable memo manifest to a file instead of applying it")
	cmd.Flags().BoolVar(&input.AnnotateResource, "annotate-resource", false, "Patch lightweight KubeMemo annotations onto the target resource")
	var expiresAt string
	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "Expiry timestamp in RFC3339")
	addOutputFlag(cmd, &output)
	_ = cmd.MarkFlagRequired("title")
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if expiresAt != "" {
			parsed, err := time.Parse(time.RFC3339, expiresAt)
			if err != nil {
				return err
			}
			input.ExpiresAt = &parsed
		}
		return nil
	}
	return cmd
}

func newSetCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace, title, expiresAt, outputPath string
	var summary, content string
	var tags []string
	var runtime bool
	input := service.UpdateNoteInput{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update a memo",
		RunE: func(cmd *cobra.Command, args []string) error {
			input.Title = title
			if cmd.Flags().Changed("summary") {
				input.Summary = &summary
			}
			if cmd.Flags().Changed("content") {
				input.Content = &content
			}
			if cmd.Flags().Changed("tag") {
				input.Tags = tags
			}
			if expiresAt != "" {
				parsed, err := time.Parse(time.RFC3339, expiresAt)
				if err != nil {
					return err
				}
				input.ExpiresAt = &parsed
			}
			input.Runtime = runtime
			input.RuntimeNamespace = runtimeNamespace
			input.OutputPath = outputPath
			result, err := opts.service.SetNote(context.Background(), input)
			if err != nil {
				return err
			}
			return writeOutput(output, result)
		},
	}
	cmd.Flags().StringVar(&input.ID, "id", "", "Memo ID")
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&summary, "summary", "", "New summary")
	cmd.Flags().StringVar(&content, "content", "", "New content")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "New tags")
	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "New expiry timestamp")
	cmd.Flags().BoolVar(&runtime, "runtime", false, "Limit update to runtime memos")
	cmd.Flags().StringVar(&outputPath, "output-path", "", "Write durable memo manifest to a file instead of applying it")
	cmd.Flags().BoolVar(&input.AnnotateResource, "annotate-resource", false, "Patch lightweight KubeMemo annotations onto the target resource")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newRemoveCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace, id string
	var expiredRuntimeOnly, removeAnnotations bool
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a memo",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.service.RemoveNote(context.Background(), id, expiredRuntimeOnly, runtimeNamespace, removeAnnotations)
			if err != nil {
				return err
			}
			return writeOutput(output, result)
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Memo ID")
	cmd.Flags().BoolVar(&expiredRuntimeOnly, "expired-runtime-only", false, "Delete expired runtime memos")
	cmd.Flags().BoolVar(&removeAnnotations, "remove-annotations", false, "Refresh target KubeMemo annotations after deleting memos")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newExportCmd(opts *rootOptions) *cobra.Command {
	var output, path, runtimeNamespace string
	var includeRuntime bool
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export memos to files",
		RunE: func(cmd *cobra.Command, args []string) error {
			written, err := opts.service.Export(context.Background(), path, includeRuntime, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeOutput(output, map[string]any{"written": written})
		},
	}
	cmd.Flags().StringVar(&path, "path", opts.cfg.GitOpsRepoPath, "Export path")
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", false, "Include runtime memos")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newImportCmd(opts *rootOptions) *cobra.Command {
	var output, path string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import memos from disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			applied, err := opts.service.Import(context.Background(), path)
			if err != nil {
				return err
			}
			return writeOutput(output, map[string]any{"applied": applied})
		},
	}
	cmd.Flags().StringVar(&path, "path", opts.cfg.GitOpsRepoPath, "Import path")
	addOutputFlag(cmd, &output)
	return cmd
}

func newSyncCmd(opts *rootOptions) *cobra.Command {
	var output, path, direction, runtimeNamespace string
	var includeRuntime bool
	cmd := &cobra.Command{
		Use:   "sync-gitops",
		Short: "Run GitOps import or export",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.service.SyncGitOps(context.Background(), path, direction, includeRuntime, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeOutput(output, result)
		},
	}
	cmd.Flags().StringVar(&path, "path", opts.cfg.GitOpsRepoPath, "GitOps path")
	cmd.Flags().StringVar(&direction, "direction", "export", "Direction: export or import")
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", false, "Include runtime memos")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newTestGitOpsCmd(opts *rootOptions) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "test-gitops-mode",
		Short: "Detect GitOps mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeOutput(output, opts.service.TestGitOpsMode(context.Background()))
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newTestRuntimeStoreCmd(opts *rootOptions) *cobra.Command {
	var output, namespace string
	cmd := &cobra.Command{
		Use:   "test-runtime-store",
		Short: "Validate the runtime store",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeOutput(output, opts.service.TestRuntimeStore(context.Background(), namespace))
		},
	}
	cmd.Flags().StringVar(&namespace, "namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newClearCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace string
	var expiredOnly bool
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear runtime memos",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.service.ClearRuntime(context.Background(), expiredOnly, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeOutput(output, result)
		},
	}
	cmd.Flags().BoolVar(&expiredOnly, "expired-only", false, "Delete only expired runtime memos")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newActivityCmd(opts *rootOptions) *cobra.Command {
	var output, kind, namespace, name, runtimeNamespace string
	cmd := &cobra.Command{
		Use:   "get-activity",
		Short: "Get runtime activity memos",
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := opts.service.GetActivity(context.Background(), kind, namespace, name, runtimeNamespace)
			if err != nil {
				return err
			}
			return writeOutput(output, model.NoteList{Items: notes})
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Target kind")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Target namespace")
	cmd.Flags().StringVar(&name, "name", "", "Target name")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newWatchActivityCmd(opts *rootOptions) *cobra.Command {
	var output, runtimeNamespace string
	var namespaces, kinds []string
	cmd := &cobra.Command{
		Use:   "watch-activity",
		Short: "Watch Kubernetes resources and auto-capture runtime activity memos",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if output != "" && output != "text" {
				return fmt.Errorf("watch-activity currently supports only text output")
			}

			fmt.Println("KubeMemo activity capture started. Press Ctrl+C to stop.")
			return opts.service.StartActivityCapture(ctx, runtimeNamespace, namespaces, kinds, func(evt service.ActivityEvent) {
				status := "captured"
				if evt.Deduplicated {
					status = "deduplicated"
				}
				fmt.Printf("[%s] %s %s %s %s -> %s\n", status, evt.Target.Kind, evt.Target.Namespace, evt.Target.Name, evt.OldValue, evt.NewValue)
			})
		},
	}
	cmd.Flags().StringSliceVar(&namespaces, "namespace", nil, "Namespace scope to watch; empty watches all accessible namespaces")
	cmd.Flags().StringSliceVar(&kinds, "kind", nil, "Kinds to watch; defaults to configured activity kinds")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	addOutputFlag(cmd, &output)
	return cmd
}

func newConfigCmd(opts *rootOptions) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get-config",
		Short: "Get the effective configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeOutput(output, opts.cfg)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}
