package plugin

import (
	"fmt"
	"github.com/dshills/alas/internal/runtime"
)

// PluginLoader interface for loading different types of plugins
type PluginLoader interface {
	Load(plugin *Plugin) error
	Unload(plugin *Plugin) error
	Call(plugin *Plugin, function string, args []runtime.Value) (runtime.Value, error)
}

// BuiltinFunctionRegistry manages built-in functions provided by plugins
type BuiltinFunctionRegistry struct {
	functions map[string]BuiltinFunction
}

// BuiltinFunction represents a native function provided by a plugin
type BuiltinFunction interface {
	Name() string
	Module() string
	Call(args []runtime.Value) (runtime.Value, error)
	Signature() FunctionDef
}

// NewBuiltinFunctionRegistry creates a new built-in function registry
func NewBuiltinFunctionRegistry() *BuiltinFunctionRegistry {
	return &BuiltinFunctionRegistry{
		functions: make(map[string]BuiltinFunction),
	}
}

// Register registers a built-in function
func (r *BuiltinFunctionRegistry) Register(fn BuiltinFunction) error {
	key := fmt.Sprintf("%s.%s", fn.Module(), fn.Name())
	if _, exists := r.functions[key]; exists {
		return fmt.Errorf("function %s already registered", key)
	}
	r.functions[key] = fn
	return nil
}

// Unregister removes a built-in function
func (r *BuiltinFunctionRegistry) Unregister(module, name string) {
	key := fmt.Sprintf("%s.%s", module, name)
	delete(r.functions, key)
}

// Get retrieves a built-in function
func (r *BuiltinFunctionRegistry) Get(module, name string) (BuiltinFunction, bool) {
	key := fmt.Sprintf("%s.%s", module, name)
	fn, exists := r.functions[key]
	return fn, exists
}

// List returns all registered functions
func (r *BuiltinFunctionRegistry) List() []BuiltinFunction {
	functions := make([]BuiltinFunction, 0, len(r.functions))
	for _, fn := range r.functions {
		functions = append(functions, fn)
	}
	return functions
}

// ModuleLoader loads ALaS module plugins
type ModuleLoader struct {
	moduleLoader ModuleLoaderFunc
}

// ModuleLoaderFunc is a function type for loading ALaS modules
type ModuleLoaderFunc func(name string) (interface{}, error)

// NewModuleLoader creates a new module loader
func NewModuleLoader(moduleLoader ModuleLoaderFunc) *ModuleLoader {
	return &ModuleLoader{
		moduleLoader: moduleLoader,
	}
}

// Load loads an ALaS module plugin
func (l *ModuleLoader) Load(plugin *Plugin) error {
	if plugin.Manifest.Type != PluginTypeModule {
		return fmt.Errorf("module loader can only load module plugins")
	}

	// Load the ALaS module file
	modulePath := fmt.Sprintf("%s/%s.alas.json", plugin.Path, plugin.Manifest.Module)
	_, err := l.moduleLoader(modulePath)
	if err != nil {
		return fmt.Errorf("failed to load module %s: %w", plugin.Manifest.Module, err)
	}

	return nil
}

// Unload unloads an ALaS module plugin
func (l *ModuleLoader) Unload(plugin *Plugin) error {
	// ALaS modules don't need explicit unloading
	return nil
}

// Call calls a function in an ALaS module plugin
func (l *ModuleLoader) Call(plugin *Plugin, function string, args []runtime.Value) (runtime.Value, error) {
	// This would delegate to the ALaS interpreter to call the function
	// For now, return an error indicating this needs interpreter integration
	return runtime.Value{}, fmt.Errorf("module function calls require interpreter integration")
}

// NativeLoader loads native shared library plugins
type NativeLoader struct {
	// Platform-specific native library loading would go here
}

// NewNativeLoader creates a new native loader
func NewNativeLoader() *NativeLoader {
	return &NativeLoader{}
}

// Load loads a native plugin
func (l *NativeLoader) Load(plugin *Plugin) error {
	if plugin.Manifest.Type != PluginTypeNative {
		return fmt.Errorf("native loader can only load native plugins")
	}

	// TODO: Implement native library loading
	// This would use dlopen/LoadLibrary depending on platform
	return fmt.Errorf("native plugin loading not yet implemented")
}

// Unload unloads a native plugin
func (l *NativeLoader) Unload(plugin *Plugin) error {
	// TODO: Implement native library unloading
	return fmt.Errorf("native plugin unloading not yet implemented")
}

// Call calls a function in a native plugin
func (l *NativeLoader) Call(plugin *Plugin, function string, args []runtime.Value) (runtime.Value, error) {
	// TODO: Implement native function calling
	return runtime.Value{}, fmt.Errorf("native plugin calls not yet implemented")
}

// HybridLoader loads hybrid plugins (ALaS modules with native functions)
type HybridLoader struct {
	moduleLoader *ModuleLoader
	nativeLoader *NativeLoader
	builtinRegistry *BuiltinFunctionRegistry
}

// NewHybridLoader creates a new hybrid loader
func NewHybridLoader(moduleLoader *ModuleLoader, nativeLoader *NativeLoader, builtinRegistry *BuiltinFunctionRegistry) *HybridLoader {
	return &HybridLoader{
		moduleLoader: moduleLoader,
		nativeLoader: nativeLoader,
		builtinRegistry: builtinRegistry,
	}
}

// Load loads a hybrid plugin
func (l *HybridLoader) Load(plugin *Plugin) error {
	if plugin.Manifest.Type != PluginTypeHybrid {
		return fmt.Errorf("hybrid loader can only load hybrid plugins")
	}

	// Load the ALaS module part
	if err := l.moduleLoader.Load(plugin); err != nil {
		return fmt.Errorf("failed to load module part: %w", err)
	}

	// Load the native part if it exists
	if len(plugin.Manifest.Implementation.Binaries) > 0 {
		if err := l.nativeLoader.Load(plugin); err != nil {
			return fmt.Errorf("failed to load native part: %w", err)
		}
	}

	// Register any native functions as built-ins
	for _, fn := range plugin.Manifest.Functions {
		if fn.Native {
			// TODO: Create built-in function wrapper
			// builtinFn := createNativeWrapper(plugin, fn)
			// l.builtinRegistry.Register(builtinFn)
		}
	}

	return nil
}

// Unload unloads a hybrid plugin
func (l *HybridLoader) Unload(plugin *Plugin) error {
	// Unregister native functions
	for _, fn := range plugin.Manifest.Functions {
		if fn.Native {
			l.builtinRegistry.Unregister(plugin.Manifest.Module, fn.Name)
		}
	}

	// Unload native part
	if len(plugin.Manifest.Implementation.Binaries) > 0 {
		if err := l.nativeLoader.Unload(plugin); err != nil {
			return fmt.Errorf("failed to unload native part: %w", err)
		}
	}

	// Unload module part
	if err := l.moduleLoader.Unload(plugin); err != nil {
		return fmt.Errorf("failed to unload module part: %w", err)
	}

	return nil
}

// Call calls a function in a hybrid plugin
func (l *HybridLoader) Call(plugin *Plugin, function string, args []runtime.Value) (runtime.Value, error) {
	// Check if it's a native function
	if fn, exists := plugin.Manifest.GetFunction(function); exists && fn.Native {
		if builtinFn, exists := l.builtinRegistry.Get(plugin.Manifest.Module, function); exists {
			return builtinFn.Call(args)
		}
		return runtime.Value{}, fmt.Errorf("native function %s not found in registry", function)
	}

	// Otherwise delegate to module loader
	return l.moduleLoader.Call(plugin, function, args)
}