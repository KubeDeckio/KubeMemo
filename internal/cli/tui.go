package cli

import (
	"context"

	"github.com/KubeDeckio/KubeMemo/internal/tui"
	"github.com/spf13/cobra"
)

func newTuiCmd(opts *rootOptions) *cobra.Command {
	includeRuntime := true
	var runtimeNamespace string
	var namespaces []string
	var autoRefreshSeconds int
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the KubeMemo terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run(context.Background(), opts.service, tui.Options{
				IncludeRuntime:     includeRuntime,
				RuntimeNamespace:   runtimeNamespace,
				Namespaces:         namespaces,
				AutoRefreshSeconds: autoRefreshSeconds,
			})
		},
	}
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", true, "Include runtime memos")
	cmd.Flags().StringVar(&runtimeNamespace, "runtime-namespace", opts.cfg.RuntimeNamespace, "Runtime namespace")
	cmd.Flags().StringSliceVar(&namespaces, "namespace", nil, "Namespace scope")
	cmd.Flags().IntVar(&autoRefreshSeconds, "auto-refresh-seconds", 3, "Refresh interval in seconds")
	return cmd
}
