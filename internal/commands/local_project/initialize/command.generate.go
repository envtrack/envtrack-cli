package initialize

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/envtrack/envtrack-cli/internal/commandconfig"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

// CmdGenerateCommand returns a Cobra command that builds a command string from
// a named command definition, optional session name and supplied params.
func CmdGenerateCommand() *cobra.Command {
	var session string
	var startScreen bool
	var screenName string
	cmd := &cobra.Command{
		Use:   "generate <commandName> [param1=value param2=value ...]",
		Short: "Generate a command string from a named command definition",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runGenerate(cmd, args, session, startScreen, screenName)
		},
	}
	cmd.Flags().StringVarP(&session, "session", "s", "", "Session name to apply session-specific params")
	cmd.Flags().BoolVar(&startScreen, "screen", false, "Start the generated command in a new GNU screen session if there are no warnings/errors")
	cmd.Flags().StringVar(&screenName, "screen-name", "", "Optional name for the new screen session (defaults to envtrack-<command>-<ts>)")
	return cmd
}

func runGenerate(_ *cobra.Command, args []string, session string, startScreen bool, screenName string) {
	commandName := args[0]
	supplied := make(map[string]string)
	for _, a := range args[1:] {
		if strings.Contains(a, "=") {
			parts := strings.SplitN(a, "=", 2)
			supplied[parts[0]] = parts[1]
		} else {
			fmt.Printf("Invalid param '%s', expected key=value\n", a)
			return
		}
	}

	localCfgParams, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfgParams == nil {
		fmt.Printf("Error loading local config: %v\n", err)
		return
	}

	if localCfgParams.CommandsConfiguration == nil || localCfgParams.CommandsConfiguration.ExternalConfigPath == nil {
		fmt.Println("No commands configuration set in local config.")
		return
	}

	commandsConfiguration, err := commandconfig.LoadFromFile(*localCfgParams.CommandsConfiguration.ExternalConfigPath)
	if err != nil {
		fmt.Printf("Error loading commands configuration from file: %v\n", err)
		return
	}

	parsedCommands, pcErr := commandconfig.ParseInternalCommandsFromCommandConfig(commandsConfiguration)
	if pcErr != nil {
		fmt.Printf("Warning: error parsing commands: %v\n", pcErr)
	}

	ic, ok := parsedCommands[commandName]
	if !ok {
		fmt.Printf("Command '%s' not found in configuration\n", commandName)
		return
	}

	// If a session was provided, check it exists and apply its values to overwrite the command's params and configMap
	if session != "" {
		if ic.Sessions == nil {
			fmt.Printf("Warning: session '%s' specified but command has no sessions defined; ignoring session\n", session)
			// clear session to avoid further session-specific lookups
			session = ""
		} else {
			if sc, sok := ic.Sessions[session]; sok {
				// merge session params into command params (override)
				if sc.Params != nil {
					if ic.Params == nil {
						ic.Params = make(map[string]string)
					}
					for k, v := range sc.Params {
						ic.Params[k] = v
					}
				}
				// merge session configMap into command configMap (override)
				if sc.ConfigMap != nil {
					if ic.ConfigMap == nil {
						ic.ConfigMap = make(map[string]string)
					}
					for k, v := range sc.ConfigMap {
						ic.ConfigMap[k] = v
					}
				}
				// we've merged the session into the command; clear the session name to avoid double-application downstream
				session = ""
			} else {
				fmt.Printf("Warning: session '%s' not found for command '%s'; ignoring session\n", session, commandName)
				session = ""
			}
		}
	}

	// Pass the full commandsConfiguration to BuildCommandString (signature updated)
	cmdStr, warnings, berr := commandconfig.BuildCommandString(commandsConfiguration, ic, session, supplied)
	// Print any warnings
	for _, w := range warnings {
		fmt.Printf("Warning: %s\n", w)
	}
	if berr != nil {
		fmt.Printf("Error building command string: %v\n", berr)
		return
	}

	fmt.Println(cmdStr)

	// If requested, start the generated command in a new GNU screen session, but only if there were no warnings
	if startScreen {
		if len(warnings) > 0 {
			fmt.Println("Not starting screen because there are warnings. Resolve warnings first or run without --screen to just print the command.")
			return
		}

		// determine screen session name
		if screenName == "" {
			screenName = fmt.Sprintf("envtrack-%s-%d", commandName, time.Now().Unix())
		}

		// Print current user for diagnostics
		if cur, uerr := user.Current(); uerr == nil {
			fmt.Printf("launching screen as user: %s (uid=%s gid=%s)\n", cur.Username, cur.Uid, cur.Gid)
		} else {
			fmt.Printf("unable to determine current user: %v\n", uerr)
		}

		// Ensure 'screen' binary exists in PATH and resolve its absolute path
		screenPath, lerr := exec.LookPath("screen")
		if lerr != nil {
			fmt.Printf("screen not found in PATH: %v\n", lerr)
			return
		}
		fmt.Printf("using screen binary: %s\n", screenPath)

		// Prefer bash when available to ensure 'exec bash' keeps the session interactive
		shellPath := "sh"
		if bp, berr := exec.LookPath("bash"); berr == nil {
			shellPath = bp
			fmt.Printf("using shell: %s\n", shellPath)
		} else {
			fmt.Printf("bash not found, falling back to sh\n")
		}

		// Keep the screen session alive after the command finishes by dropping to a shell.
		// Also redirect command stdout/stderr to a logfile in /tmp for diagnostics.
		logPath := fmt.Sprintf("/tmp/%s.log", screenName)
		fullCmd := fmt.Sprintf("%s > %s 2>&1; exec %s", cmdStr, logPath, shellPath)
		scrCmd := exec.Command(screenPath, "-dmS", screenName, shellPath, "-lc", fullCmd)
		// Capture combined output for diagnostics
		out, err := scrCmd.CombinedOutput()
		// Print full output for diagnostics regardless of error to help troubleshooting
		if len(out) > 0 {
			fmt.Printf("screen output: %s\n", string(out))
		}
		if err != nil {
			fmt.Printf("Error starting screen session: %v; output: %s\n", err, string(out))
			return
		}
		// After starting, list sessions to show sockets and session ids
		lsOut, lsErr := exec.Command(screenPath, "-ls").CombinedOutput()
		if len(lsOut) > 0 {
			fmt.Printf("screen -ls output: %s\n", string(lsOut))
		}
		if lsErr != nil {
			fmt.Printf("Warning: error listing screen sessions: %v\n", lsErr)
		}
		fmt.Printf("Started screen session '%s' (log: %s)\n", screenName, logPath)
	}
}
