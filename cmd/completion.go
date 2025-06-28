package cmd

import (
	"bytes"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:
  $ source <(gman completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ gman completion bash > /etc/bash_completion.d/gman
  # macOS:
  $ gman completion bash > /usr/local/etc/bash_completion.d/gman

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ gman completion zsh > "${fpath[1]}/_gman"

  # You will need to start a new shell for this setup to take effect.

fish:
  $ gman completion fish | source

  # To load completions for each session, execute once:
  $ gman completion fish > ~/.config/fish/completions/gman.fish

PowerShell:
  PS> gman completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> gman completion powershell > gman.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			// Handle zsh completion with cleaning to remove "-- 4" artifact
			var buf bytes.Buffer
			cmd.Root().GenZshCompletion(&buf)
			
			// Clean the output by removing the "-- 4" directive output
			cleanedOutput := strings.ReplaceAll(buf.String(), " -- 4", "")
			cleanedOutput = strings.ReplaceAll(cleanedOutput, "-- 4", "")
			
			os.Stdout.WriteString(cleanedOutput)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
