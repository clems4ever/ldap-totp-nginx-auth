package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script",
		Long:                  completionLong,
		Args:                  cobra.ExactValidArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		Run:                   cmdCompletionRun,
	}

	return cmd
}

func cmdCompletionRun(cmd *cobra.Command, args []string) {
	var err error

	switch args[0] {
	case "bash":
		err = cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		err = cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		err = cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		fmt.Printf("Invalid shell provided for completion command: %s\n", args[0])
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error generating completion: %v\n", err)
		os.Exit(1)
	}
}
