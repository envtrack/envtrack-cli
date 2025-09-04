// ...existing code...
package initialize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/envtrack/envtrack-cli/internal/commandconfig"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

// Flags for the web debug command
var (
	webPort                int
	webHost                string
	hideWebCommandConfig   bool
	hideWebParsedConfigs   bool
	hideWebParsedTemplates bool
	hideWebParsedCommands  bool
)

// CmdDebugWebCommand starts a small HTTP server that serves an HTML page
// and exposes the same JSON that cmd-debug prints via a WebSocket endpoint.
func CmdDebugWebCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start a web server and show parsed commands configuration in the browser (WebSocket)",
		Run:   runCmdDebugWeb,
	}

	cmd.Flags().IntVarP(&webPort, "port", "P", 8080, "Port to listen on")
	cmd.Flags().StringVarP(&webHost, "host", "H", "localhost", "Host to bind to")

	// same selection flags but separate variables to avoid interfering with CLI-only command
	cmd.Flags().BoolVarP(&hideWebCommandConfig, "no-command-config", "c", false, "Do not include commandConfig in output")
	cmd.Flags().BoolVarP(&hideWebParsedConfigs, "no-parsed-configs", "p", false, "Do not include parsedConfigs in output")
	cmd.Flags().BoolVarP(&hideWebParsedTemplates, "no-parsed-templates", "t", false, "Do not include parsedTemplates in output")
	cmd.Flags().BoolVarP(&hideWebParsedCommands, "no-parsed-commands", "m", false, "Do not include parsedCommands in output")

	return cmd
}

func buildDebugOutput(hideCmd, hideCfgs, hideTemps bool) ([]byte, error) {
	localCfgParams, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfgParams == nil {
		return nil, fmt.Errorf("Error loading local config: %v", err)
	}

	if localCfgParams.CommandsConfiguration == nil {
		return nil, fmt.Errorf("No commands configuration set in local config.")
	}

	commandsConfiguration, err := commandconfig.LoadFromFile(*localCfgParams.CommandsConfiguration.ExternalConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading commands configuration from file: %v", err)
	}

	parsedConfigs, pErr := commandconfig.ParseConfigsFromCommandConfig(commandsConfiguration)
	if pErr != nil {
		// Create a ParsedConfigs compatible warning entry instead of using a generic map
		parsedConfigs = make(commandconfig.ParsedConfigs)
		parsedConfigs["_parseWarning"] = map[string]interface{}{"value": pErr.Error()}
	}

	parsedTemplates, tErr := commandconfig.ParseTemplatesFromCommandConfig(commandsConfiguration)
	if tErr != nil {
		//parsedTemplates = make(commandconfig.ParsedTemplates)
		//parsedTemplates["_parseWarning"] = map[string]interface{}{"value": tErr.Error()}
	}

	// Parse typed InternalCommand objects
	parsedCommands, pcErr := commandconfig.ParseInternalCommandsFromCommandConfig(commandsConfiguration)
	if pcErr != nil {
		fmt.Printf("Warning: error parsing commands: %v\n", pcErr)
	}

	outObj := map[string]interface{}{}
	if !hideCmd {
		outObj["commandConfig"] = commandsConfiguration
	}
	if !hideCfgs {
		outObj["parsedConfigs"] = parsedConfigs
	}
	if !hideTemps {
		outObj["parsedTemplates"] = parsedTemplates
	}
	if !hideParsedCommands {
		outObj["parsedCommands"] = parsedCommands
	}

	if len(outObj) == 0 {
		return nil, fmt.Errorf("No output selected (commandConfig, parsedConfigs and parsedTemplates are hidden). Use flags to include output.")
	}

	// Produce a stable, canonical JSON representation so repeated marshals of maps
	// don't result in different key orders and thus different bytes.
	var generic interface{}
	bRaw, err := json.Marshal(outObj)
	if err != nil {
		return nil, fmt.Errorf("Error preparing debug output: %v", err)
	}
	if err := json.Unmarshal(bRaw, &generic); err != nil {
		return nil, fmt.Errorf("Error preparing debug output: %v", err)
	}
	out, merr := stableMarshal(generic)
	if merr != nil {
		return nil, fmt.Errorf("Error serializing local commands configuration: %v", merr)
	}

	return out, nil
}

// stableMarshal marshals generic JSON-compatible data (maps, slices, primitives)
// using deterministic key ordering for maps. Output is indented for readability.
func stableMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeValue(&buf, v, 0); err != nil {
		return nil, err
	}
	// pretty-print with indentation
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, buf.Bytes(), "", "  "); err != nil {
		return nil, err
	}
	return pretty.Bytes(), nil
}

func writeValue(buf *bytes.Buffer, v interface{}, _depth int) error {
	switch t := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, _ := json.Marshal(k)
			buf.Write(kb)
			buf.WriteByte(':')
			if err := writeValue(buf, t[k], _depth+1); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case []interface{}:
		buf.WriteByte('[')
		for i, el := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeValue(buf, el, _depth+1); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case string:
		b, _ := json.Marshal(t)
		buf.Write(b)
	case float64, bool, nil:
		b, _ := json.Marshal(t)
		buf.Write(b)
	default:
		// Fallback: marshal using encoding/json
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}
		buf.Write(b)
	}
	return nil
}

func runCmdDebugWeb(_ *cobra.Command, _ []string) {
	addr := webHost + ":" + strconv.Itoa(webPort)

	// Handlers
	html := getWebHtml()

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})

	http.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		data, err := buildDebugOutput(hideWebCommandConfig, hideWebParsedConfigs, hideWebParsedTemplates)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(data)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		data, err := buildDebugOutput(hideWebCommandConfig, hideWebParsedConfigs, hideWebParsedTemplates)
		if err != nil {
			// send error as text message and close
			_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, data)
		// keep connection open and periodically send updates every 5s
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				data, err := buildDebugOutput(hideWebCommandConfig, hideWebParsedConfigs, hideWebParsedTemplates)
				if err != nil {
					_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			}
		}
	})

	// Start server with graceful shutdown
	srv := &http.Server{Addr: addr}

	go func() {
		fmt.Printf("Serving debug web UI at http://%s/\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	fmt.Println("Server stopped")
}

// Prevent unused function warning
var _ = CmdDebugWebCommand
