package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for llm-imager.

To load completions:

Bash:
  $ source <(llm-imager completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ llm-imager completion bash > /etc/bash_completion.d/llm-imager
  # macOS:
  $ llm-imager completion bash > $(brew --prefix)/etc/bash_completion.d/llm-imager

Zsh:
  $ source <(llm-imager completion zsh)
  # To load completions for each session, execute once:
  $ llm-imager completion zsh > "${fpath[1]}/_llm-imager"

Fish:
  $ llm-imager completion fish | source
  # To load completions for each session, execute once:
  $ llm-imager completion fish > ~/.config/fish/completions/llm-imager.fish

PowerShell:
  PS> llm-imager completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, add to profile:
  PS> llm-imager completion powershell >> $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
