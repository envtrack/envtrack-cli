package initialize

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/envtrack/envtrack-cli/internal/commandconfig"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

// CmdWalkCommand returns a Cobra command that walks the parsed configuration objects
// and prints a human friendly tree representation. It can also print a specific
// nested object as JSON via the --path flag. Use --interactive to open an
// interactive tree viewer (TUI).
func CmdWalkCommand() *cobra.Command {
	var path string
	var asJSON bool
	var interactive bool

	cmd := &cobra.Command{
		Use:   "walk",
		Short: "Walk parsed command configuration objects",
		Run: func(cmd *cobra.Command, _ []string) {
			runCmdWalk(cmd, path, asJSON, interactive)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Optional path to a specific object to print (format: <configKey>[.nested.key])")
	cmd.Flags().BoolVar(&asJSON, "json", false, "When a specific path is selected, print the object as JSON")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Open an interactive tree viewer for parsed configs")
	return cmd
}

func runCmdWalk(_ *cobra.Command, path string, asJSON bool, interactive bool) {
	localCfgParams, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfgParams == nil {
		fmt.Printf("Error loading local config: %v\n", err)
		return
	}

	if localCfgParams.CommandsConfiguration == nil {
		fmt.Println("No commands configuration set in local config.")
		return
	}

	commandsConfiguration, err := commandconfig.LoadFromFile(*localCfgParams.CommandsConfiguration.ExternalConfigPath)
	if err != nil {
		fmt.Printf("Error loading commands configuration from file: %v\n", err)
		return
	}

	// Otherwise, if the config is embedded in the local config, print it and parse any referenced configs.
	parsedConfigs, pErr := commandconfig.ParseConfigsFromCommandConfig(commandsConfiguration)
	if pErr != nil {
		fmt.Printf("Warning: error parsing referenced configs: %v\n", pErr)
	}

	if interactive {
		runInteractive(parsedConfigs)
		return
	}

	if path == "" {
		// Print all top-level parsed configs
		for k, v := range parsedConfigs {
			fmt.Println(k)
			printMap(v, 1)
			fmt.Println()
		}
		return
	}

	// Parse path: split into config key and optional nested path separated by '.'
	cfgKey := path
	nested := ""
	if idx := strings.Index(path, "."); idx >= 0 {
		cfgKey = path[:idx]
		nested = path[idx+1:]
	}

	root, ok := parsedConfigs[cfgKey]
	if !ok {
		fmt.Printf("Config key not found: %s\n", cfgKey)
		return
	}

	var selected interface{} = root
	if nested != "" {
		parts := strings.Split(nested, ".")
		cur := interface{}(root)
		for _, p := range parts {
			if m, ok := cur.(map[string]interface{}); ok {
				if val, exists := m[p]; exists {
					cur = val
					continue
				}
				fmt.Printf("Path not found: %s\n", p)
				return
			}
			fmt.Printf("Cannot descend into non-object at %s\n", p)
			return
		}
		selected = cur
	}

	if asJSON {
		out, jerr := json.MarshalIndent(selected, "", "  ")
		if jerr != nil {
			fmt.Printf("Error serializing object as JSON: %v\n", jerr)
			return
		}
		fmt.Println(string(out))
		return
	}

	// Print selected object as tree
	switch t := selected.(type) {
	case map[string]interface{}:
		printMap(t, 0)
	case []interface{}:
		printArray(t, 0)
	default:
		fmt.Printf("%v\n", t)
	}
}

// runInteractive opens a TUI tree viewer using tview. Basic navigation:
// - arrows / j/k: move
// - Enter: toggle expand/collapse
// - q or Esc: quit
// Selected node details are shown in the right pane as JSON.
func runInteractive(parsed commandconfig.ParsedConfigs) {
	app := tview.NewApplication()

	rootNode := tview.NewTreeNode("Parsed Configs").SetColor(tcell.ColorGreen)
	// ensure stable ordering
	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		n := tview.NewTreeNode(k).SetReference(parsed[k])
		rootNode.AddChild(n)
		// fully populate children so search/navigation can rely on node structure
		populateTreeNode(n, parsed[k])
	}

	tree := tview.NewTreeView().SetRoot(rootNode).SetCurrentNode(rootNode)
	tree.SetBorder(true).SetTitle("Configs")

	detail := tview.NewTextView().SetDynamicColors(true)
	detail.SetBorder(true).SetTitle("Details")
	detail.SetChangedFunc(func() { app.Draw() })

	// search state
	var matches []*tview.TreeNode
	currentMatch := -1

	// helper to select match
	selectMatch := func(idx int) {
		if len(matches) == 0 || idx < 0 || idx >= len(matches) {
			return
		}
		node := matches[idx]
		// ensure ancestors are expanded so node is visible
		expandAncestors(rootNode, node)
		tree.SetCurrentNode(node)
		showNodeDetail(detail, node)
		currentMatch = idx
	}

	// show details of current node when changed/selected
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if node.IsExpanded() {
			node.Collapse()
		} else {
			node.Expand()
		}
		showNodeDetail(detail, node)
	})

	// update detail when selection moves
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		showNodeDetail(detail, node)
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex := tview.NewFlex().AddItem(tree, 0, 1, true).AddItem(detail, 0, 2, false)
	// status bar
	status := tview.NewTextView().SetDynamicColors(true)
	status.SetText("Press / to search, n/N to navigate matches, q to quit")
	status.SetBorder(false)

	flex.AddItem(mainFlex, 0, 1, true)
	flex.AddItem(status, 1, 0, false)

	pages := tview.NewPages()
	pages.AddPage("main", flex, true, true)

	// search form (overlay)
	searchForm := tview.NewForm().AddInputField("Search", "", 40, nil, nil).
		AddButton("Find", nil).
		AddButton("Cancel", nil)
	searchForm.SetBorder(true).SetTitle("Search")

	// wire buttons and input
	searchForm.GetButton(0).SetSelectedFunc(func() { // Find
		input := searchForm.GetFormItemByLabel("Search").(*tview.InputField).GetText()
		query := strings.TrimSpace(input)
		pages.HidePage("search")
		if query == "" {
			matches = nil
			currentMatch = -1
			status.SetText("No query provided. Press / to search, q to quit")
			return
		}
		// perform search
		matches = findMatchingNodes(rootNode, strings.ToLower(query))
		if len(matches) == 0 {
			status.SetText(fmt.Sprintf("No matches for '%s' — press / to search again", query))
			currentMatch = -1
			app.SetFocus(tree)
			return
		}
		currentMatch = 0
		selectMatch(currentMatch)
		status.SetText(fmt.Sprintf("%d matches for '%s' — n/N to navigate", len(matches), query))
		app.SetFocus(tree)
	})
	searchForm.GetButton(1).SetSelectedFunc(func() { // Cancel
		pages.HidePage("search")
		app.SetFocus(tree)
	})

	// pages.AddPage("search",
	//	tview.NewModalForm("", searchForm), true, false)
	// Use a Flex wrapper for the form instead of the non-existent NewModalForm helper
	pages.AddPage("search",
		tview.NewFlex().AddItem(searchForm, 0, 1, true), true, false)

	// global key handlers
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			if pages.HasPage("search") {
				pages.HidePage("search")
				app.SetFocus(tree)
				return nil
			}
			app.Stop()
			return nil
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
		case '/':
			pages.ShowPage("search")
			// focus input field
			app.SetFocus(searchForm.GetFormItem(0))
			return nil
		case 'n':
			if len(matches) > 0 {
				currentMatch = (currentMatch + 1) % len(matches)
				selectMatch(currentMatch)
				return nil
			}
		case 'N':
			if len(matches) > 0 {
				currentMatch = (currentMatch - 1 + len(matches)) % len(matches)
				selectMatch(currentMatch)
				return nil
			}
		}
		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
	}
}

// findMatchingNodes traverses the tree and returns nodes where either the node text
// or the JSON serialization of the node's reference contains the query (case-insensitive).
func findMatchingNodes(root *tview.TreeNode, query string) []*tview.TreeNode {
	var matches []*tview.TreeNode
	var walk func(n *tview.TreeNode)
	walk = func(n *tview.TreeNode) {
		// check node text
		text := strings.ToLower(n.GetText())
		if strings.Contains(text, query) {
			matches = append(matches, n)
		}
		// check reference as JSON
		if ref := n.GetReference(); ref != nil {
			if b, err := json.Marshal(ref); err == nil {
				if strings.Contains(strings.ToLower(string(b)), query) {
					matches = append(matches, n)
				}
			}
		}
		for _, c := range n.GetChildren() {
			walk(c)
		}
	}
	walk(root)
	return matches
}

// expandAncestors expands nodes from the root down to make target visible.
func expandAncestors(root, target *tview.TreeNode) {
	// find path from root to target by DFS and expand along the way
	var path []*tview.TreeNode
	var found bool
	var dfs func(n *tview.TreeNode) bool
	dfs = func(n *tview.TreeNode) bool {
		if n == target {
			path = append(path, n)
			return true
		}
		for _, c := range n.GetChildren() {
			if dfs(c) {
				path = append(path, n)
				return true
			}
		}
		return false
	}
	found = dfs(root)
	if !found {
		return
	}
	// path contains nodes from target up to root; expand from root down
	for i := len(path) - 1; i >= 0; i-- {
		path[i].Expand()
	}
}

// populateTreeNode now fully recurses to build the visible tree (no placeholders)
func populateTreeNode(parent *tview.TreeNode, v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		// stable ordering
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := tview.NewTreeNode(k).SetReference(t[k])
			parent.AddChild(child)
			populateTreeNode(child, t[k])
		}
	case []interface{}:
		for i, item := range t {
			label := fmt.Sprintf("[%d]", i)
			child := tview.NewTreeNode(label).SetReference(item)
			parent.AddChild(child)
			populateTreeNode(child, item)
		}
	default:
		// scalar; show as leaf
		parent.SetText(fmt.Sprintf("%s: %v", parent.GetText(), t))
	}
}

func showNodeDetail(detail *tview.TextView, node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		detail.SetText("<no data>")
		return
	}
	out, err := json.MarshalIndent(reference, "", "  ")
	if err != nil {
		detail.SetText(fmt.Sprintf("<error serializing: %v>", err))
		return
	}
	detail.SetText(string(out))
	// if node has unexpanded placeholder children, populate on demand
	if len(node.GetChildren()) == 1 && node.GetChildren()[0].GetText() == "" {
		// clear placeholder and populate children
		node.ClearChildren()
		if m, ok := reference.(map[string]interface{}); ok {
			populateTreeNode(node, m)
		} else if a, ok := reference.([]interface{}); ok {
			populateTreeNode(node, a)
		}
	}
}

func printMap(m map[string]interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	// Ensure stable ordering by collecting keys (simple approach)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		fmt.Printf("%s%s: ", prefix, k)
		switch t := v.(type) {
		case map[string]interface{}:
			fmt.Println()
			printMap(t, indent+1)
		case []interface{}:
			fmt.Println()
			printArray(t, indent+1)
		default:
			fmt.Printf("%v\n", t)
		}
	}
}

func printArray(a []interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	for i, v := range a {
		fmt.Printf("%s- [%d] ", prefix, i)
		switch t := v.(type) {
		case map[string]interface{}:
			fmt.Println()
			printMap(t, indent+1)
		case []interface{}:
			fmt.Println()
			printArray(t, indent+1)
		default:
			fmt.Printf("%v\n", t)
		}
	}
}
