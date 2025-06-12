package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest defines the structure of a plugin manifest file
type Manifest struct {
	// Plugin metadata
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`

	// Plugin type and capabilities
	Type         PluginType   `json:"type"`
	Capabilities []Capability `json:"capabilities"`

	// Module and function definitions
	Module    string            `json:"module"`
	Functions []FunctionDef     `json:"functions"`
	Types     []TypeDef         `json:"types,omitempty"`
	
	// Dependencies and compatibility
	AlasVersion   string            `json:"alas_version"`
	Dependencies  []string          `json:"dependencies,omitempty"`
	
	// Implementation details
	Implementation Implementation `json:"implementation"`
	
	// Security and runtime settings
	Security SecurityPolicy `json:"security"`
	Runtime  RuntimeConfig  `json:"runtime"`
}

// PluginType defines the type of plugin
type PluginType string

const (
	PluginTypeNative   PluginType = "native"   // Compiled shared library
	PluginTypeModule   PluginType = "module"   // Pure ALaS module
	PluginTypeHybrid   PluginType = "hybrid"   // ALaS module with native functions
	PluginTypeBuiltin  PluginType = "builtin"  // Built into the runtime
)

// Capability defines what the plugin can do
type Capability string

const (
	CapabilityFunction    Capability = "function"     // Provides functions
	CapabilityType        Capability = "type"         // Provides custom types
	CapabilityModule      Capability = "module"       // Provides modules
	CapabilityCodegen     Capability = "codegen"      // Extends code generation
	CapabilityValidation  Capability = "validation"   // Extends validation
	CapabilityIO          Capability = "io"           // Performs I/O operations
	CapabilityNetwork     Capability = "network"      // Network access
	CapabilityFileSystem  Capability = "filesystem"   // File system access
	CapabilityProcess     Capability = "process"      // Process execution
)

// FunctionDef defines a function provided by the plugin
type FunctionDef struct {
	Name        string      `json:"name"`
	Params      []ParamDef  `json:"params"`
	Returns     string      `json:"returns"`
	Description string      `json:"description"`
	Native      bool        `json:"native,omitempty"`
	Async       bool        `json:"async,omitempty"`
}

// ParamDef defines a function parameter
type ParamDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TypeDef defines a custom type provided by the plugin
type TypeDef struct {
	Name        string                 `json:"name"`
	Kind        string                 `json:"kind"` // struct, enum, alias
	Fields      []FieldDef             `json:"fields,omitempty"`
	Values      []string               `json:"values,omitempty"`
	BaseType    string                 `json:"base_type,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// FieldDef defines a field in a struct type
type FieldDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Implementation defines how the plugin is implemented
type Implementation struct {
	Language   string            `json:"language"`   // go, rust, c, alas
	EntryPoint string            `json:"entrypoint"` // main function or shared library path
	BuildCmd   string            `json:"build_cmd,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	Binaries   []string          `json:"binaries,omitempty"`
	Config     map[string]string `json:"config,omitempty"`
}

// SecurityPolicy defines security constraints for the plugin
type SecurityPolicy struct {
	Sandbox     bool     `json:"sandbox"`
	AllowedAPIs []string `json:"allowed_apis,omitempty"`
	MaxMemory   string   `json:"max_memory,omitempty"`
	MaxCPU      string   `json:"max_cpu,omitempty"`
	Timeout     string   `json:"timeout,omitempty"`
}

// RuntimeConfig defines runtime behavior
type RuntimeConfig struct {
	Lazy        bool              `json:"lazy"`         // Load on first use
	Persistent  bool              `json:"persistent"`   // Keep loaded between calls
	Parallel    bool              `json:"parallel"`     // Allow parallel execution
	Environment map[string]string `json:"environment,omitempty"`
}

// LoadManifest loads a plugin manifest from a file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	// Set default values
	if manifest.Type == "" {
		manifest.Type = PluginTypeModule
	}
	
	if manifest.AlasVersion == "" {
		manifest.AlasVersion = ">=0.1.0"
	}

	return &manifest, nil
}

// SaveManifest saves a plugin manifest to a file
func (m *Manifest) SaveManifest(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// Validate checks if the manifest is valid
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	
	if m.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	
	if m.Module == "" {
		return fmt.Errorf("plugin module name is required")
	}

	// Validate plugin type
	switch m.Type {
	case PluginTypeNative, PluginTypeModule, PluginTypeHybrid, PluginTypeBuiltin:
		// Valid
	default:
		return fmt.Errorf("invalid plugin type: %s", m.Type)
	}

	// Validate capabilities
	for _, cap := range m.Capabilities {
		switch cap {
		case CapabilityFunction, CapabilityType, CapabilityModule, CapabilityCodegen,
			 CapabilityValidation, CapabilityIO, CapabilityNetwork, CapabilityFileSystem,
			 CapabilityProcess:
			// Valid
		default:
			return fmt.Errorf("invalid capability: %s", cap)
		}
	}

	// Validate functions
	for _, fn := range m.Functions {
		if fn.Name == "" {
			return fmt.Errorf("function name is required")
		}
		if fn.Returns == "" {
			return fmt.Errorf("function return type is required for %s", fn.Name)
		}
	}

	// Validate implementation
	if m.Implementation.Language == "" {
		return fmt.Errorf("implementation language is required")
	}

	return nil
}

// HasCapability checks if the plugin has a specific capability
func (m *Manifest) HasCapability(cap Capability) bool {
	for _, c := range m.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// GetFunction returns a function definition by name
func (m *Manifest) GetFunction(name string) (*FunctionDef, bool) {
	for _, fn := range m.Functions {
		if fn.Name == name {
			return &fn, true
		}
	}
	return nil, false
}