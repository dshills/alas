package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/dshills/alas/internal/plugin"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		listCommand()
	case "info":
		infoCommand()
	case "install":
		installCommand()
	case "uninstall":
		uninstallCommand()
	case "create":
		createCommand()
	case "validate":
		validateCommand()
	case "load":
		loadCommand()
	case "unload":
		unloadCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ALaS Plugin Manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  alas-plugin <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                    List all discovered plugins")
	fmt.Println("  info <plugin>           Show detailed plugin information")
	fmt.Println("  install <path>          Install plugin from path")
	fmt.Println("  uninstall <plugin>      Uninstall plugin")
	fmt.Println("  create <name>           Create new plugin template")
	fmt.Println("  validate <path>         Validate plugin manifest")
	fmt.Println("  load <plugin>           Load plugin")
	fmt.Println("  unload <plugin>         Unload plugin")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -path <path>            Plugin search path (default: ./plugins)")
	fmt.Println("  -format <format>        Output format: table, json (default: table)")
}

func getRegistry(searchPath string) *plugin.Registry {
	registry := plugin.NewRegistry()

	if searchPath == "" {
		searchPath = "./plugins"
	}

	registry.AddSearchPath(searchPath)

	// Discover plugins
	if err := registry.Discover(); err != nil {
		fmt.Printf("Error discovering plugins: %v\n", err)
		os.Exit(1)
	}

	return registry
}

func listCommand() {
	var formatFlag string
	var pathFlag string

	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.StringVar(&formatFlag, "format", "table", "Output format (table, json)")
	fs.StringVar(&pathFlag, "path", "./plugins", "Plugin search path")
	fs.Parse(os.Args[2:])

	registry := getRegistry(pathFlag)
	plugins := registry.List()

	if formatFlag == "json" {
		printPluginsJSON(plugins)
	} else {
		printPluginsTable(plugins)
	}
}

func printPluginsTable(plugins []*plugin.Plugin) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tTYPE\tSTATE\tMODULE\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t----\t-----\t------\t-----------")

	for _, p := range plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			p.Manifest.Name,
			p.Manifest.Version,
			p.Manifest.Type,
			p.State,
			p.Manifest.Module,
			truncateString(p.Manifest.Description, 50))
	}

	w.Flush()
}

func printPluginsJSON(plugins []*plugin.Plugin) {
	data, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func infoCommand() {
	var pathFlag string

	fs := flag.NewFlagSet("info", flag.ExitOnError)
	fs.StringVar(&pathFlag, "path", "./plugins", "Plugin search path")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 1 {
		fmt.Println("Usage: alas-plugin info <plugin>")
		os.Exit(1)
	}

	pluginName := args[0]
	registry := getRegistry(pathFlag)

	p, exists := registry.Get(pluginName)
	if !exists {
		fmt.Printf("Plugin '%s' not found\n", pluginName)
		os.Exit(1)
	}

	printPluginInfo(p)
}

func printPluginInfo(p *plugin.Plugin) {
	fmt.Printf("Plugin: %s\n", p.Manifest.Name)
	fmt.Printf("Version: %s\n", p.Manifest.Version)
	fmt.Printf("Type: %s\n", p.Manifest.Type)
	fmt.Printf("State: %s\n", p.State)
	fmt.Printf("Module: %s\n", p.Manifest.Module)
	fmt.Printf("Description: %s\n", p.Manifest.Description)
	fmt.Printf("Author: %s\n", p.Manifest.Author)
	fmt.Printf("License: %s\n", p.Manifest.License)
	fmt.Printf("Path: %s\n", p.Path)
	fmt.Printf("ALaS Version: %s\n", p.Manifest.AlasVersion)

	if len(p.Manifest.Capabilities) > 0 {
		fmt.Printf("Capabilities: %s\n", strings.Join(capabilitiesToStrings(p.Manifest.Capabilities), ", "))
	}

	if len(p.Manifest.Dependencies) > 0 {
		fmt.Printf("Dependencies: %s\n", strings.Join(p.Manifest.Dependencies, ", "))
	}

	if len(p.Manifest.Functions) > 0 {
		fmt.Println("\nFunctions:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  NAME\tPARAMS\tRETURNS\tNATIVE\tDESCRIPTION")
		fmt.Fprintln(w, "  ----\t------\t-------\t------\t-----------")

		for _, fn := range p.Manifest.Functions {
			params := make([]string, len(fn.Params))
			for i, param := range fn.Params {
				params[i] = fmt.Sprintf("%s:%s", param.Name, param.Type)
			}

			fmt.Fprintf(w, "  %s\t%s\t%s\t%v\t%s\n",
				fn.Name,
				strings.Join(params, ", "),
				fn.Returns,
				fn.Native,
				truncateString(fn.Description, 40))
		}
		w.Flush()
	}

	fmt.Printf("\nImplementation:\n")
	fmt.Printf("  Language: %s\n", p.Manifest.Implementation.Language)
	fmt.Printf("  Entry Point: %s\n", p.Manifest.Implementation.EntryPoint)

	fmt.Printf("\nSecurity:\n")
	fmt.Printf("  Sandbox: %v\n", p.Manifest.Security.Sandbox)
	if p.Manifest.Security.MaxMemory != "" {
		fmt.Printf("  Max Memory: %s\n", p.Manifest.Security.MaxMemory)
	}
	if p.Manifest.Security.Timeout != "" {
		fmt.Printf("  Timeout: %s\n", p.Manifest.Security.Timeout)
	}
}

func installCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: alas-plugin install <path>")
		os.Exit(1)
	}

	sourcePath := os.Args[2]

	// TODO: Implement plugin installation
	// This would copy the plugin to the plugins directory and register it
	fmt.Printf("Installing plugin from %s...\n", sourcePath)
	fmt.Println("Plugin installation not yet implemented")
}

func uninstallCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: alas-plugin uninstall <plugin>")
		os.Exit(1)
	}

	pluginName := os.Args[2]

	// TODO: Implement plugin uninstallation
	fmt.Printf("Uninstalling plugin %s...\n", pluginName)
	fmt.Println("Plugin uninstallation not yet implemented")
}

func createCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: alas-plugin create <name>")
		os.Exit(1)
	}

	pluginName := os.Args[2]

	// Create plugin directory
	pluginDir := filepath.Join(".", pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		fmt.Printf("Error creating plugin directory: %v\n", err)
		os.Exit(1)
	}

	// Create manifest template
	manifest := &plugin.Manifest{
		Name:         pluginName,
		Version:      "0.1.0",
		Description:  fmt.Sprintf("ALaS plugin: %s", pluginName),
		Author:       "Your Name",
		License:      "MIT",
		Type:         plugin.PluginTypeModule,
		Capabilities: []plugin.Capability{plugin.CapabilityFunction},
		Module:       pluginName,
		Functions: []plugin.FunctionDef{
			{
				Name:        "hello",
				Params:      []plugin.ParamDef{{Name: "name", Type: "string"}},
				Returns:     "string",
				Description: "Example function that greets someone",
			},
		},
		AlasVersion: ">=0.1.0",
		Implementation: plugin.Implementation{
			Language:   "alas",
			EntryPoint: pluginName + ".alas.json",
		},
		Security: plugin.SecurityPolicy{
			Sandbox: true,
		},
		Runtime: plugin.RuntimeConfig{
			Lazy:       true,
			Persistent: false,
		},
	}

	// Save manifest
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if err := manifest.SaveManifest(manifestPath); err != nil {
		fmt.Printf("Error saving manifest: %v\n", err)
		os.Exit(1)
	}

	// Create example ALaS module
	modulePath := filepath.Join(pluginDir, pluginName+".alas.json")
	moduleContent := fmt.Sprintf(`{
  "type": "module",
  "name": "%s",
  "exports": ["hello"],
  "functions": [
    {
      "type": "function",
      "name": "hello",
      "params": [
        {
          "name": "name",
          "type": "string"
        }
      ],
      "returns": "string",
      "body": [
        {
          "type": "return",
          "value": {
            "type": "binary",
            "op": "+",
            "left": {
              "type": "literal",
              "value": "Hello, "
            },
            "right": {
              "type": "variable",
              "name": "name"
            }
          }
        }
      ]
    }
  ]
}`, pluginName)

	if err := os.WriteFile(modulePath, []byte(moduleContent), 0600); err != nil {
		fmt.Printf("Error writing module file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created plugin template in %s/\n", pluginDir)
	fmt.Printf("  - plugin.json (manifest)\n")
	fmt.Printf("  - %s.alas.json (module)\n", pluginName)
}

func validateCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: alas-plugin validate <path>")
		os.Exit(1)
	}

	manifestPath := os.Args[2]

	manifest, err := plugin.LoadManifest(manifestPath)
	if err != nil {
		fmt.Printf("Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	if err := manifest.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Plugin manifest is valid\n")
}

func pluginOperation(operation string, action func(*plugin.Registry, string) error) {
	var pathFlag string

	fs := flag.NewFlagSet(operation, flag.ExitOnError)
	fs.StringVar(&pathFlag, "path", "./plugins", "Plugin search path")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 1 {
		fmt.Printf("Usage: alas-plugin %s <plugin>\n", operation)
		os.Exit(1)
	}

	pluginName := args[0]
	registry := getRegistry(pathFlag)

	if err := action(registry, pluginName); err != nil {
		fmt.Printf("Error %sing plugin: %v\n", operation, err)
		os.Exit(1)
	}

	fmt.Printf("Plugin %s %sed successfully\n", pluginName, operation)
}

func loadCommand() {
	pluginOperation("load", func(registry *plugin.Registry, name string) error {
		return registry.Load(name)
	})
}

func unloadCommand() {
	pluginOperation("unload", func(registry *plugin.Registry, name string) error {
		return registry.Unload(name)
	})
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func capabilitiesToStrings(caps []plugin.Capability) []string {
	result := make([]string, len(caps))
	for i, cap := range caps {
		result[i] = string(cap)
	}
	return result
}
