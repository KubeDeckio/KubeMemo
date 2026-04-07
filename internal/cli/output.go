package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type outputMode string

const (
	outputText outputMode = "text"
	outputJSON outputMode = "json"
	outputYAML outputMode = "yaml"
)

func addOutputFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVar(target, "output", string(outputText), "Output format: text, json, yaml")
}

func writeOutput(mode string, value any) error {
	switch outputMode(mode) {
	case outputJSON:
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	case outputYAML:
		data, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	default:
		switch v := value.(type) {
		case string:
			fmt.Println(v)
		default:
			data, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		}
		return nil
	}
}
