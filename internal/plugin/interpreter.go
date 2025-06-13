package plugin

import (
	"fmt"

	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
)

// PluginAwareInterpreter extends the base interpreter with plugin support.
type PluginAwareInterpreter struct { //nolint:revive // Name is intentional for clarity
	*interpreter.Interpreter
	registry        *Registry
	builtinRegistry *BuiltinFunctionRegistry
}

// NewPluginAwareInterpreter creates a new plugin-aware interpreter.
func NewPluginAwareInterpreter(base *interpreter.Interpreter, registry *Registry) *PluginAwareInterpreter {
	return &PluginAwareInterpreter{
		Interpreter:     base,
		registry:        registry,
		builtinRegistry: NewBuiltinFunctionRegistry(),
	}
}

// CallBuiltinFunction calls a built-in function by name with pre-evaluated arguments.
func (i *PluginAwareInterpreter) CallBuiltinFunction(functionName string, args []runtime.Value) (runtime.Value, error) {
	// Parse module.function format
	module, function, err := parseBuiltinName(functionName)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("invalid builtin function name %s: %w", functionName, err)
	}

	// Try to find the function in the builtin registry first
	if builtinFn, exists := i.builtinRegistry.Get(module, function); exists {
		return builtinFn.Call(args)
	}

	// Try to find the function in plugins
	plugin, fnDef, err := i.registry.GetFunction(module, function)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("builtin function %s not found: %w", functionName, err)
	}

	// Validate argument count
	if len(args) != len(fnDef.Params) {
		return runtime.NewVoid(), fmt.Errorf("function %s expects %d arguments, got %d",
			functionName, len(fnDef.Params), len(args))
	}

	// Call the plugin function
	if plugin.Loader != nil {
		return plugin.Loader.Call(plugin, function, args)
	}

	return runtime.NewVoid(), fmt.Errorf("plugin %s not loaded", plugin.Manifest.Name)
}

// RegisterBuiltinFunction registers a builtin function.
func (i *PluginAwareInterpreter) RegisterBuiltinFunction(fn BuiltinFunction) error {
	return i.builtinRegistry.Register(fn)
}

// UnregisterBuiltinFunction unregisters a builtin function.
func (i *PluginAwareInterpreter) UnregisterBuiltinFunction(module, name string) {
	i.builtinRegistry.Unregister(module, name)
}

// GetBuiltinFunction retrieves a builtin function.
func (i *PluginAwareInterpreter) GetBuiltinFunction(module, name string) (BuiltinFunction, bool) {
	return i.builtinRegistry.Get(module, name)
}

// LoadPlugin loads a plugin and registers its functions.
func (i *PluginAwareInterpreter) LoadPlugin(name string) error {
	return i.registry.Load(name)
}

// UnloadPlugin unloads a plugin and unregisters its functions.
func (i *PluginAwareInterpreter) UnloadPlugin(name string) error {
	plugin, exists := i.registry.Get(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Unregister builtin functions
	for _, fn := range plugin.Manifest.Functions {
		if fn.Native {
			i.builtinRegistry.Unregister(plugin.Manifest.Module, fn.Name)
		}
	}

	return i.registry.Unload(name)
}

// GetRegistry returns the plugin registry.
func (i *PluginAwareInterpreter) GetRegistry() *Registry {
	return i.registry
}

// parseBuiltinName parses a builtin function name in module.function format.
func parseBuiltinName(name string) (module, function string, err error) {
	// Find the last dot to separate module from function
	lastDot := -1
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 {
		return "", "", fmt.Errorf("builtin function name must be in module.function format")
	}

	module = name[:lastDot]
	function = name[lastDot+1:]

	if module == "" || function == "" {
		return "", "", fmt.Errorf("both module and function name must be non-empty")
	}

	return module, function, nil
}

// InterpreterPluginManager manages the integration between the interpreter and plugins.
type InterpreterPluginManager struct {
	interpreter *PluginAwareInterpreter
	registry    *Registry
}

// NewInterpreterPluginManager creates a new interpreter plugin manager.
func NewInterpreterPluginManager(base *interpreter.Interpreter) *InterpreterPluginManager {
	registry := NewRegistry()
	pluginInterpreter := NewPluginAwareInterpreter(base, registry)

	return &InterpreterPluginManager{
		interpreter: pluginInterpreter,
		registry:    registry,
	}
}

// Initialize sets up the plugin system.
func (m *InterpreterPluginManager) Initialize(pluginPaths []string) error {
	// Add search paths
	for _, path := range pluginPaths {
		m.registry.AddSearchPath(path)
	}

	// Register default loaders
	moduleLoader := NewModuleLoader(func(name string) (interface{}, error) {
		// This would integrate with the base interpreter's module loading
		return nil, fmt.Errorf("module loading integration needed")
	})

	nativeLoader := NewNativeLoader()
	hybridLoader := NewHybridLoader(moduleLoader, nativeLoader, m.interpreter.builtinRegistry)

	m.registry.RegisterLoader(PluginTypeModule, moduleLoader)
	m.registry.RegisterLoader(PluginTypeNative, nativeLoader)
	m.registry.RegisterLoader(PluginTypeHybrid, hybridLoader)

	// Discover plugins
	if err := m.registry.Discover(); err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}

	return nil
}

// GetInterpreter returns the plugin-aware interpreter.
func (m *InterpreterPluginManager) GetInterpreter() *PluginAwareInterpreter {
	return m.interpreter
}

// GetRegistry returns the plugin registry.
func (m *InterpreterPluginManager) GetRegistry() *Registry {
	return m.registry
}
