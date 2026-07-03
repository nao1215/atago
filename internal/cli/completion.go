package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// subcommandNames is the stable, sorted list of top-level subcommands that shell
// completion offers. It is the single source of truth for the generated scripts,
// so adding or removing a subcommand deterministically updates completion output
// (and the golden tests that guard it).
var subcommandNames = []string{
	"completion",
	"doc",
	"explain",
	"help",
	"init",
	"list",
	"manifest",
	"run",
	"snapshot",
	"version",
}

// runFlags is the stable, sorted list of `atago run` flags surfaced to
// completion. Keep it in sync with runCmd's FlagSet.
var runFlags = []string{
	"--artifacts-dir",
	"--ci",
	"--fail-fast",
	"--filter",
	"--parallel",
	"--repeat",
	"--report",
	"--rerun-failed",
	"--retry-failed",
	"--skip-tag",
	"--tag",
	"--update-snapshots",
	"--verbose",
}

// completionShells is the sorted set of shells `atago completion` supports.
var completionShells = []string{"bash", "fish", "powershell", "zsh"}

// completionCmd implements `atago completion <bash|zsh|fish|powershell>` (#62):
// print a deterministic completion script for the requested shell. The scripts
// are versioned by the runtime surface (subcommandNames / runFlags) so a change
// to the CLI surface is a visible diff in the golden tests.
func completionCmd(args []string, stdout, stderr io.Writer) int {
	// --help must behave like every other subcommand's --help, not be
	// mistaken for a shell name.
	if wantsHelp(args) {
		fmt.Fprintf(stdout, "Usage: atago completion <%s>\n", strings.Join(completionShells, "|"))
		return ExitOK
	}
	if len(args) != 1 {
		fmt.Fprintf(stderr, "Usage: atago completion <%s>\n", strings.Join(completionShells, "|"))
		return ExitConfig
	}
	shell := args[0]
	script, ok := completionScript(shell)
	if !ok {
		fmt.Fprintf(stderr, "atago completion: unknown shell %q (want %s)\n", shell, strings.Join(completionShells, ", "))
		return ExitConfig
	}
	fmt.Fprint(stdout, script)
	return ExitOK
}

// completionScript returns the completion script for shell and whether shell is
// supported.
func completionScript(shell string) (string, bool) {
	switch shell {
	case "bash":
		return bashCompletion(), true
	case "zsh":
		return zshCompletion(), true
	case "fish":
		return fishCompletion(), true
	case "powershell":
		return powershellCompletion(), true
	default:
		return "", false
	}
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func bashCompletion() string {
	cmds := strings.Join(sortedCopy(subcommandNames), " ")
	flags := strings.Join(sortedCopy(runFlags), " ")
	return fmt.Sprintf(`# bash completion for atago
# Install: atago completion bash > /etc/bash_completion.d/atago
#      or: source <(atago completion bash)
_atago() {
    local cur prev words cword
    _init_completion 2>/dev/null || {
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    }
    local commands="%s"
    local run_flags="%s"
    if [ "${COMP_CWORD}" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
        return 0
    fi
    case "${COMP_WORDS[1]}" in
        run)
            COMPREPLY=( $(compgen -W "${run_flags}" -- "${cur}") $(compgen -f -- "${cur}") )
            ;;
        *)
            COMPREPLY=( $(compgen -f -- "${cur}") )
            ;;
    esac
    return 0
}
complete -F _atago atago
`, cmds, flags)
}

func zshCompletion() string {
	cmds := strings.Join(sortedCopy(subcommandNames), " ")
	flags := strings.Join(sortedCopy(runFlags), " ")
	return fmt.Sprintf(`#compdef atago
# zsh completion for atago
# Install: atago completion zsh > "${fpath[1]}/_atago"
_atago() {
    local -a commands run_flags
    commands=(%s)
    run_flags=(%s)
    if (( CURRENT == 2 )); then
        compadd -- ${commands}
        return
    fi
    case ${words[2]} in
        run)
            compadd -- ${run_flags}
            _files
            ;;
        *)
            _files
            ;;
    esac
}
compdef _atago atago
`, cmds, flags)
}

func fishCompletion() string {
	var b strings.Builder
	b.WriteString("# fish completion for atago\n")
	b.WriteString("# Install: atago completion fish > ~/.config/fish/completions/atago.fish\n")
	b.WriteString("complete -c atago -f\n")
	for _, c := range sortedCopy(subcommandNames) {
		fmt.Fprintf(&b, "complete -c atago -n '__fish_use_subcommand' -a %s\n", c)
	}
	for _, f := range sortedCopy(runFlags) {
		long := strings.TrimPrefix(f, "--")
		fmt.Fprintf(&b, "complete -c atago -n '__fish_seen_subcommand_from run' -l %s\n", long)
	}
	return b.String()
}

func powershellCompletion() string {
	cmds := strings.Join(quoteAll(sortedCopy(subcommandNames)), ", ")
	flags := strings.Join(quoteAll(sortedCopy(runFlags)), ", ")
	return fmt.Sprintf(`# PowerShell completion for atago
# Install: atago completion powershell | Out-String | Invoke-Expression
Register-ArgumentCompleter -Native -CommandName atago -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @(%s)
    $runFlags = @(%s)
    $tokens = $commandAst.CommandElements
    if ($tokens.Count -le 2) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }
    if ($tokens[1].Value -eq 'run') {
        $runFlags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
        }
    }
}
`, cmds, flags)
}

func quoteAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = "'" + s + "'"
	}
	return out
}
