package plugin

import (
	"fmt"
	"path/filepath"
	"sort"
	"sync"
)

// Plugin represents a loaded plugin
type Plugin struct {
	Manifest *Manifest
	Path     string
	State    PluginState
	Loader   PluginLoader
}

// PluginState represents the current state of a plugin
type PluginState string

const (
	StateUnloaded PluginState = "unloaded"
	StateLoading  PluginState = "loading"
	StateLoaded   PluginState = "loaded"
	StateError    PluginState = "error"
)

// Registry manages all plugins in the system
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
	paths   []string
	loaders map[PluginType]PluginLoader
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
		paths:   make([]string, 0),
		loaders: make(map[PluginType]PluginLoader),
	}
}

// AddSearchPath adds a directory to search for plugins
func (r *Registry) AddSearchPath(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, path)
}

// RegisterLoader registers a plugin loader for a specific plugin type
func (r *Registry) RegisterLoader(pluginType PluginType, loader PluginLoader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loaders[pluginType] = loader
}

// Discover scans search paths for plugins and loads their manifests
func (r *Registry) Discover() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, searchPath := range r.paths {
		if err := r.discoverInPath(searchPath); err != nil {
			return fmt.Errorf("failed to discover plugins in %s: %w", searchPath, err)
		}
	}

	return nil
}

// discoverInPath scans a single directory for plugins
func (r *Registry) discoverInPath(searchPath string) error {
	matches, err := filepath.Glob(filepath.Join(searchPath, "*", "plugin.json"))
	if err != nil {
		return err
	}

	for _, manifestPath := range matches {
		pluginDir := filepath.Dir(manifestPath)
		
		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			// Log error but continue discovering other plugins
			continue
		}

		if err := manifest.Validate(); err != nil {
			// Log validation error but continue
			continue
		}

		// Check if plugin already exists
		if _, exists := r.plugins[manifest.Name]; exists {
			// Log duplicate plugin warning
			continue
		}

		plugin := &Plugin{
			Manifest: manifest,
			Path:     pluginDir,
			State:    StateUnloaded,
		}

		r.plugins[manifest.Name] = plugin
	}

	return nil
}

// Register manually registers a plugin
func (r *Registry) Register(manifest *Manifest, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid plugin manifest: %w", err)
	}

	if _, exists := r.plugins[manifest.Name]; exists {
		return fmt.Errorf("plugin %s already registered", manifest.Name)
	}

	plugin := &Plugin{
		Manifest: manifest,
		Path:     path,
		State:    StateUnloaded,
	}

	r.plugins[manifest.Name] = plugin
	return nil
}

// Load loads a plugin by name
func (r *Registry) Load(name string) error {
	r.mu.Lock()
	plugin, exists := r.plugins[name]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("plugin %s not found", name)
	}
	r.mu.Unlock()

	return r.loadPlugin(plugin)
}

// LoadAll loads all discovered plugins
func (r *Registry) LoadAll() error {
	r.mu.RLock()
	plugins := make([]*Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	r.mu.RUnlock()

	for _, plugin := range plugins {
		if err := r.loadPlugin(plugin); err != nil {
			// Log error but continue loading other plugins
			plugin.State = StateError
		}
	}

	return nil
}

// loadPlugin loads a single plugin
func (r *Registry) loadPlugin(plugin *Plugin) error {
	if plugin.State == StateLoaded {
		return nil
	}

	plugin.State = StateLoading

	loader, exists := r.loaders[plugin.Manifest.Type]
	if !exists {
		plugin.State = StateError
		return fmt.Errorf("no loader registered for plugin type %s", plugin.Manifest.Type)
	}

	if err := loader.Load(plugin); err != nil {
		plugin.State = StateError
		return fmt.Errorf("failed to load plugin %s: %w", plugin.Manifest.Name, err)
	}

	plugin.Loader = loader
	plugin.State = StateLoaded
	return nil
}

// Unload unloads a plugin by name
func (r *Registry) Unload(name string) error {
	r.mu.Lock()
	plugin, exists := r.plugins[name]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("plugin %s not found", name)
	}
	r.mu.Unlock()

	if plugin.State != StateLoaded {
		return nil
	}

	if plugin.Loader != nil {
		if err := plugin.Loader.Unload(plugin); err != nil {
			return fmt.Errorf("failed to unload plugin %s: %w", name, err)
		}
	}

	plugin.State = StateUnloaded
	plugin.Loader = nil
	return nil
}

// Get returns a plugin by name
func (r *Registry) Get(name string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all registered plugins
func (r *Registry) List() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	// Sort by name for consistent ordering
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Manifest.Name < plugins[j].Manifest.Name
	})

	return plugins
}

// ListByCapability returns plugins that have a specific capability
func (r *Registry) ListByCapability(cap Capability) []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*Plugin, 0)
	for _, plugin := range r.plugins {
		if plugin.Manifest.HasCapability(cap) {
			plugins = append(plugins, plugin)
		}
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Manifest.Name < plugins[j].Manifest.Name
	})

	return plugins
}

// GetFunction finds a function across all loaded plugins
func (r *Registry) GetFunction(module, name string) (*Plugin, *FunctionDef, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, plugin := range r.plugins {
		if plugin.State != StateLoaded {
			continue
		}

		if plugin.Manifest.Module == module {
			if fn, exists := plugin.Manifest.GetFunction(name); exists {
				return plugin, fn, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("function %s.%s not found in any loaded plugin", module, name)
}

// HasModule checks if a module is provided by any loaded plugin
func (r *Registry) HasModule(module string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, plugin := range r.plugins {
		if plugin.State == StateLoaded && plugin.Manifest.Module == module {
			return true
		}
	}

	return false
}

// GetStats returns registry statistics
func (r *Registry) GetStats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RegistryStats{
		SearchPaths: len(r.paths),
		Loaders:     len(r.loaders),
	}

	for _, plugin := range r.plugins {
		stats.TotalPlugins++
		switch plugin.State {
		case StateLoaded:
			stats.LoadedPlugins++
		case StateError:
			stats.ErrorPlugins++
		}
	}

	return stats
}

// RegistryStats contains statistics about the plugin registry
type RegistryStats struct {
	SearchPaths   int
	Loaders       int
	TotalPlugins  int
	LoadedPlugins int
	ErrorPlugins  int
}