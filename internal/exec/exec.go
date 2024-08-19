package exec

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/alessio/shellescape" // For safe shell escaping
)

type Command struct {
	Name         string `json:"name"`
	Command      string `json:"command"`
	Description  string `json:"description"`
	Background   bool   `json:"background"`
	ForceRestart bool   `json:"forceRestart"`
}

type CommandManager struct {
	Commands map[string]*Command
	LogFile  string
	EnvVars  map[string]string // Store environment variables
}

func NewCommandManager(logFile string) *CommandManager {
	return &CommandManager{
		Commands: make(map[string]*Command),
		LogFile:  logFile,
		EnvVars:  make(map[string]string),
	}
}

func (cm *CommandManager) AddCommand(cmd *Command) {
	cm.Commands[cmd.Name] = cmd
}

func (cm *CommandManager) SetEnvVar(key, value string) {
	cm.EnvVars[key] = value
}

func (cm *CommandManager) ExecuteCommand(name string) error {
	cmd, ok := cm.Commands[name]
	if !ok {
		return fmt.Errorf("command %s not found", name)
	}

	expandedCommand := cm.expandVariables(cmd.Command)

	if cmd.Background {
		return cm.executeBackgroundCommand(cmd, expandedCommand)
	}
	return cm.executeForegroundCommand(cmd, expandedCommand)
}

func (cm *CommandManager) expandVariables(command string) string {
	for key, value := range cm.EnvVars {
		command = strings.ReplaceAll(command, "{{"+key+"}}", shellescape.Quote(value))
	}
	return command
}

func (cm *CommandManager) executeBackgroundCommand(cmd *Command, expandedCommand string) error {
	// Check if the command is already running
	running, pid, err := cm.isCommandRunning(expandedCommand)
	if err != nil {
		return fmt.Errorf("failed to check if command is running: %v", err)
	}

	if running {
		if cmd.ForceRestart {
			if err := cm.killProcess(pid); err != nil {
				return fmt.Errorf("failed to kill existing process: %v", err)
			}
		} else {
			return fmt.Errorf("command %s is already running", cmd.Name)
		}
	}

	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("cmd", "/C", expandedCommand)
	} else {
		execCmd = exec.Command("sh", "-c", expandedCommand)
	}

	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Log output
	go cm.logCommandOutput(cmd.Name, execCmd)

	return nil
}

func (cm *CommandManager) executeForegroundCommand(cmd *Command, expandedCommand string) error {
	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("cmd", "/C", expandedCommand)
	} else {
		execCmd = exec.Command("sh", "-c", expandedCommand)
	}

	// Set up pipes for stdout and stderr
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the command
	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Log and print output
	go cm.logAndPrintOutput(cmd.Name, stdout, false)
	go cm.logAndPrintOutput(cmd.Name, stderr, true)

	// Wait for the command to finish
	if err := execCmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	return nil
}

func (cm *CommandManager) isCommandRunning(expandedCommand string) (bool, int, error) {
	var cmdOutput []byte
	var err error
	var pid int

	if runtime.GOOS == "windows" {
		cmdOutput, err = exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", "cmd.exe"), "/FO", "CSV", "/NH").Output()
		// Note: This is a simplification. In practice, you'd need a more sophisticated
		// method to accurately identify if the specific command is running on Windows.
	} else {
		cmdOutput, err = exec.Command("pgrep", "-f", expandedCommand).Output()
		if err == nil && len(cmdOutput) > 0 {
			fmt.Sscanf(string(cmdOutput), "%d", &pid)
		}
	}

	if err != nil {
		// If the command fails, it usually means the process is not running
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, 0, nil
			}
		}
		return false, 0, err
	}

	return len(cmdOutput) > 0, pid, nil
}

func (cm *CommandManager) killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid)).Run()
	}

	return process.Signal(syscall.SIGTERM)
}

func (cm *CommandManager) logCommandOutput(name string, cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe for %s: %v\n", name, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe for %s: %v\n", name, err)
		return
	}

	go cm.logAndPrintOutput(name, stdout, false)
	go cm.logAndPrintOutput(name, stderr, true)
}

func (cm *CommandManager) logAndPrintOutput(name string, reader io.Reader, isError bool) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		cm.logToFile(name, line, isError)
		if isError {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, line)
		} else {
			fmt.Printf("%s: %s\n", name, line)
		}
	}
}

func (cm *CommandManager) logToFile(name, message string, isError bool) {
	logType := "INFO"
	if isError {
		logType = "ERROR"
	}
	logMessage := fmt.Sprintf("[%s] %s: %s\n", logType, name, message)

	f, err := os.OpenFile(cm.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(logMessage); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

func (cm *CommandManager) ExecuteMultiCommand(name string, commands []string) error {
	for _, cmdName := range commands {
		if strings.HasPrefix(cmdName, "source ") {
			// Handle sourcing environment files
			envFile := strings.TrimPrefix(cmdName, "source ")
			if err := cm.sourceEnvFile(envFile); err != nil {
				return fmt.Errorf("failed to source %s: %v", envFile, err)
			}
		} else {
			if err := cm.ExecuteCommand(cmdName); err != nil {
				return fmt.Errorf("failed to execute %s: %v", cmdName, err)
			}
		}
	}
	return nil
}

func (cm *CommandManager) sourceEnvFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		os.Setenv(key, value)
	}
	return nil
}
