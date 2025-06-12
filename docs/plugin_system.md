# ALaS Plugin System

The ALaS Plugin System provides a comprehensive framework for extending the ALaS language with custom functionality through dynamically loadable plugins. The system supports multiple plugin types and provides a rich API for plugin development.

## Overview

The plugin system is designed around these core principles:

- **Modular Architecture**: Plugins are self-contained modules with clear interfaces
- **Type Safety**: All plugin functions have complete type signatures
- **Security**: Sandbox and resource controls for plugin execution
- **Discoverability**: Automatic plugin discovery and registration
- **Extensibility**: Support for multiple plugin types and implementation languages

## Plugin Types

### Module Plugins (`module`)
Pure ALaS modules that extend functionality using the ALaS language itself.

- **Implementation**: Written in ALaS JSON format
- **Performance**: Interpreted execution
- **Security**: Full sandbox support
- **Use Cases**: Business logic, algorithms, data processing

### Native Plugins (`native`)
Compiled shared libraries that provide high-performance native functions.

- **Implementation**: C, Rust, Go, or other compiled languages
- **Performance**: Native execution speed
- **Security**: Limited sandbox support
- **Use Cases**: Performance-critical operations, system integration

### Hybrid Plugins (`hybrid`)
Combination of ALaS modules with native function implementations.

- **Implementation**: ALaS modules calling native functions
- **Performance**: Mixed interpreted/native execution
- **Security**: Configurable sandbox
- **Use Cases**: Complex plugins with both logic and performance requirements

### Built-in Plugins (`builtin`)
Plugins that are compiled into the ALaS runtime itself.

- **Implementation**: Go code integrated with the interpreter
- **Performance**: Native execution speed
- **Security**: Full system access
- **Use Cases**: Core system functionality, standard library

## Plugin Manifest

Every plugin must include a `plugin.json` manifest file that describes the plugin's metadata, capabilities, and interface.

### Basic Structure

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Example plugin",
  "author": "Plugin Author",
  "license": "MIT",
  "type": "module",
  "capabilities": ["function"],
  "module": "my_plugin",
  "functions": [...],
  "alas_version": ">=0.1.0",
  "implementation": {...},
  "security": {...},
  "runtime": {...}
}
```

### Manifest Fields

#### Metadata
- `name`: Unique plugin identifier
- `version`: Semantic version string
- `description`: Human-readable description
- `author`: Plugin author/maintainer
- `license`: License identifier (SPDX format)

#### Plugin Configuration
- `type`: Plugin type (`module`, `native`, `hybrid`, `builtin`)
- `capabilities`: Array of plugin capabilities
- `module`: ALaS module name provided by this plugin
- `alas_version`: Compatible ALaS version range
- `dependencies`: Array of required module/plugin names

#### Function Definitions
- `functions`: Array of function definitions with signatures
- `types`: Array of custom type definitions (optional)

#### Implementation Details
- `implementation.language`: Implementation language
- `implementation.entrypoint`: Main file or entry point
- `implementation.build_cmd`: Build command (for compiled plugins)
- `implementation.sources`: Source file list
- `implementation.binaries`: Binary file list
- `implementation.config`: Implementation-specific configuration

#### Security Policy
- `security.sandbox`: Enable sandboxing
- `security.allowed_apis`: Allowed API access list
- `security.max_memory`: Memory limit
- `security.max_cpu`: CPU usage limit
- `security.timeout`: Execution timeout

#### Runtime Configuration
- `runtime.lazy`: Load on first use
- `runtime.persistent`: Keep loaded between calls
- `runtime.parallel`: Allow parallel execution
- `runtime.environment`: Environment variables

### Function Definitions

```json
{
  "name": "my_function",
  "params": [
    {"name": "input", "type": "string"},
    {"name": "count", "type": "int"}
  ],
  "returns": "array",
  "description": "Function description",
  "native": false,
  "async": false
}
```

## Plugin Development

### Creating a Module Plugin

1. **Create Plugin Directory**
   ```bash
   mkdir my-plugin
   cd my-plugin
   ```

2. **Create Manifest** (`plugin.json`)
   ```json
   {
     "name": "my-plugin",
     "version": "1.0.0",
     "description": "My custom plugin",
     "type": "module",
     "module": "my_plugin",
     "functions": [
       {
         "name": "process",
         "params": [{"name": "data", "type": "string"}],
         "returns": "string",
         "description": "Process input data"
       }
     ],
     "implementation": {
       "language": "alas",
       "entrypoint": "my_plugin.alas.json"
     }
   }
   ```

3. **Implement Module** (`my_plugin.alas.json`)
   ```json
   {
     "type": "module",
     "name": "my_plugin",
     "exports": ["process"],
     "functions": [
       {
         "type": "function",
         "name": "process",
         "params": [{"name": "data", "type": "string"}],
         "returns": "string",
         "body": [
           {
             "type": "return",
             "value": {
               "type": "binary",
               "op": "+",
               "left": {"type": "literal", "value": "Processed: "},
               "right": {"type": "variable", "name": "data"}
             }
           }
         ]
       }
     ]
   }
   ```

### Creating a Hybrid Plugin

Hybrid plugins combine ALaS modules with native function implementations:

1. **Manifest with Native Functions**
   ```json
   {
     "type": "hybrid",
     "functions": [
       {
         "name": "fast_compute",
         "params": [{"name": "data", "type": "array"}],
         "returns": "float",
         "native": true,
         "description": "High-performance computation"
       },
       {
         "name": "helper",
         "params": [{"name": "x", "type": "int"}],
         "returns": "int",
         "native": false,
         "description": "ALaS helper function"
       }
     ],
     "implementation": {
       "language": "go",
       "entrypoint": "plugin.go",
       "binaries": ["plugin.so"]
     }
   }
   ```

2. **Native Implementation** (Go example)
   ```go
   package main

   import "C"
   import "github.com/dshills/alas/internal/runtime"

   //export FastCompute
   func FastCompute(args []runtime.Value) runtime.Value {
       // High-performance native implementation
       return runtime.NewFloat(42.0)
   }

   func main() {} // Required for Go plugins
   ```

## Plugin Management

### CLI Commands

The `alas-plugin` command provides comprehensive plugin management:

#### List Plugins
```bash
alas-plugin list
alas-plugin list -format json
```

#### Plugin Information
```bash
alas-plugin info my-plugin
```

#### Plugin Lifecycle
```bash
alas-plugin load my-plugin
alas-plugin unload my-plugin
```

#### Plugin Development
```bash
alas-plugin create new-plugin
alas-plugin validate plugin.json
```

#### Plugin Installation
```bash
alas-plugin install /path/to/plugin
alas-plugin uninstall my-plugin
```

### Programmatic API

```go
import "github.com/dshills/alas/internal/plugin"

// Create registry
registry := plugin.NewRegistry()
registry.AddSearchPath("./plugins")

// Discover and load plugins
registry.Discover()
registry.LoadAll()

// Use plugins with interpreter
manager := plugin.NewInterpreterPluginManager(interpreter)
manager.Initialize([]string{"./plugins"})

pluginInterpreter := manager.GetInterpreter()
```

## Plugin Discovery

The plugin system automatically discovers plugins in configured search paths:

1. **Default Paths**: `./plugins`, `~/.alas/plugins`, `/usr/local/lib/alas/plugins`
2. **Environment**: `ALAS_PLUGIN_PATH` environment variable
3. **Configuration**: Plugin paths in ALaS configuration files
4. **Command Line**: `-plugin-path` flag in ALaS tools

### Discovery Process

1. Scan search paths for directories containing `plugin.json`
2. Load and validate plugin manifests
3. Register plugins in the plugin registry
4. Resolve dependencies between plugins
5. Initialize plugin loaders based on plugin types

## Security Model

### Sandboxing

Module plugins run in a secure sandbox that:

- Limits memory and CPU usage
- Restricts file system access
- Controls network access
- Provides isolated execution environment

### Capability System

Plugins declare required capabilities in their manifest:

- `function`: Provide callable functions
- `type`: Define custom types
- `module`: Provide ALaS modules
- `io`: Perform I/O operations
- `network`: Access network resources
- `filesystem`: Access file system
- `process`: Execute external processes

### Resource Limits

```json
{
  "security": {
    "sandbox": true,
    "max_memory": "100MB",
    "max_cpu": "50%",
    "timeout": "10s",
    "allowed_apis": ["std.io", "std.math"]
  }
}
```

## Integration with ALaS

### Built-in Function Calls

Plugins integrate seamlessly with ALaS's built-in function mechanism:

```json
{
  "type": "builtin",
  "name": "my_plugin.process",
  "args": [
    {"type": "literal", "value": "input data"}
  ]
}
```

### Module System Integration

Plugin modules work with ALaS's import/export system:

```json
{
  "type": "module",
  "name": "my_program",
  "imports": ["my_plugin"],
  "functions": [
    {
      "name": "main",
      "body": [
        {
          "type": "assign",
          "target": "result",
          "value": {
            "type": "module_call",
            "module": "my_plugin",
            "name": "process",
            "args": [{"type": "literal", "value": "data"}]
          }
        }
      ]
    }
  ]
}
```

## Best Practices

### Plugin Design

1. **Single Responsibility**: Each plugin should have a focused purpose
2. **Clear Interfaces**: Use descriptive function names and types
3. **Error Handling**: Return structured error information
4. **Documentation**: Provide comprehensive function descriptions
5. **Versioning**: Use semantic versioning for compatibility

### Performance

1. **Native Functions**: Use native implementations for performance-critical code
2. **Lazy Loading**: Enable lazy loading for large plugins
3. **Resource Management**: Set appropriate resource limits
4. **Caching**: Use persistent loading for frequently used plugins

### Security

1. **Minimal Capabilities**: Request only necessary capabilities
2. **Input Validation**: Validate all input parameters
3. **Sandboxing**: Enable sandboxing for untrusted plugins
4. **Regular Updates**: Keep plugins updated for security fixes

## Examples

### Hello World Plugin

See [examples/plugins/hello-world/](../examples/plugins/hello-world/) for a complete example of a simple module plugin.

### Math Utils Plugin

See [examples/plugins/math-utils/](../examples/plugins/math-utils/) for an example of a more complex plugin with multiple functions and dependencies.

## Troubleshooting

### Common Issues

1. **Plugin Not Found**: Check search paths and manifest location
2. **Load Failures**: Verify manifest syntax and dependencies
3. **Function Errors**: Check function signatures and parameter types
4. **Permission Denied**: Review security settings and capabilities

### Debug Mode

Enable plugin debugging:

```bash
export ALAS_PLUGIN_DEBUG=1
alas-run -plugin-debug my-program.alas.json
```

### Log Files

Plugin system logs are written to:
- `~/.alas/logs/plugins.log`
- System log (when installed system-wide)

## Future Enhancements

- **Plugin Marketplace**: Central repository for plugin discovery and installation
- **Hot Reloading**: Update plugins without restarting the interpreter
- **Advanced Security**: Fine-grained permission system and code signing
- **Performance Monitoring**: Plugin performance metrics and profiling
- **IDE Integration**: Plugin development tools and debugging support