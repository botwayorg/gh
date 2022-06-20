package ghconfig

import (
	"fmt"
	"strings"

	"github.com/botwayorg/gh/core/config"
	cmdGet "github.com/botwayorg/gh/pkg/cmd/gh-config/get"
	cmdSet "github.com/botwayorg/gh/pkg/cmd/gh-config/set"
	"github.com/botwayorg/gh/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	longDoc := strings.Builder{}
	longDoc.WriteString("Display or change configuration settings for botway.\n\n")
	longDoc.WriteString("Current respected settings:\n")
	for _, co := range config.ConfigOptions() {
		longDoc.WriteString(fmt.Sprintf("- %s: %s", co.Key, co.Description))
		if co.DefaultValue != "" {
			longDoc.WriteString(fmt.Sprintf(" (default: %q)", co.DefaultValue))
		}

		longDoc.WriteRune('\n')
	}

	cmd := &cobra.Command{
		Use:   "gh-config <command>",
		Short: "Manage configuration of github for botway.",
		Long:  longDoc.String(),
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(cmdGet.NewCmdConfigGet(f, nil))
	cmd.AddCommand(cmdSet.NewCmdConfigSet(f, nil))

	return cmd
}
