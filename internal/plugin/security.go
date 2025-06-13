package plugin

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// SecurityContext provides security constraints and monitoring for plugin execution.
type SecurityContext struct {
	MaxMemory    int64         // Maximum memory usage in bytes
	MaxCPU       time.Duration // Maximum CPU time
	Timeout      time.Duration // Maximum execution time
	AllowedAPIs  []string      // Allowed API modules
	Sandbox      bool          // Enable sandboxing
	capabilities map[Capability]bool
}

// NewSecurityContext creates a new security context from a security policy.
func NewSecurityContext(policy SecurityPolicy) (*SecurityContext, error) {
	ctx := &SecurityContext{
		AllowedAPIs:  policy.AllowedAPIs,
		Sandbox:      policy.Sandbox,
		capabilities: make(map[Capability]bool),
	}

	// Parse memory limit
	if policy.MaxMemory != "" {
		memBytes, err := parseMemoryLimit(policy.MaxMemory)
		if err != nil {
			return nil, fmt.Errorf("invalid memory limit: %w", err)
		}
		ctx.MaxMemory = memBytes
	}

	// Parse CPU limit
	if policy.MaxCPU != "" {
		cpuDuration, err := time.ParseDuration(policy.MaxCPU)
		if err != nil {
			return nil, fmt.Errorf("invalid CPU limit: %w", err)
		}
		ctx.MaxCPU = cpuDuration
	}

	// Parse timeout
	if policy.Timeout != "" {
		timeout, err := time.ParseDuration(policy.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
		ctx.Timeout = timeout
	}

	return ctx, nil
}

// SetCapabilities sets the allowed capabilities for this context.
func (sc *SecurityContext) SetCapabilities(caps []Capability) {
	sc.capabilities = make(map[Capability]bool)
	for _, cap := range caps {
		sc.capabilities[cap] = true
	}
}

// HasCapability checks if a capability is allowed.
func (sc *SecurityContext) HasCapability(cap Capability) bool {
	return sc.capabilities[cap]
}

// CheckAPIAccess verifies if access to an API module is allowed.
func (sc *SecurityContext) CheckAPIAccess(module string) error {
	if len(sc.AllowedAPIs) == 0 {
		return nil // No restrictions
	}

	for _, allowed := range sc.AllowedAPIs {
		if allowed == module || allowed == "*" {
			return nil
		}
	}

	return fmt.Errorf("access to API module %s not allowed", module)
}

// ExecutionMonitor monitors plugin execution for resource usage and security violations.
type ExecutionMonitor struct {
	ctx        context.Context
	cancel     context.CancelFunc
	startTime  time.Time
	startMem   uint64
	security   *SecurityContext
	violations []string
}

// NewExecutionMonitor creates a new execution monitor.
func NewExecutionMonitor(security *SecurityContext) *ExecutionMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	if security.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), security.Timeout)
	}

	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	return &ExecutionMonitor{
		ctx:        ctx,
		cancel:     cancel,
		startTime:  time.Now(),
		startMem:   startMem.Alloc,
		security:   security,
		violations: make([]string, 0),
	}
}

// Start begins monitoring plugin execution.
func (em *ExecutionMonitor) Start() {
	if em.security.MaxMemory > 0 || em.security.MaxCPU > 0 {
		go em.monitorResources()
	}
}

// Stop stops monitoring and returns any violations.
func (em *ExecutionMonitor) Stop() []string {
	em.cancel()
	return em.violations
}

// Context returns the cancellation context.
func (em *ExecutionMonitor) Context() context.Context {
	return em.ctx
}

// monitorResources continuously monitors resource usage.
func (em *ExecutionMonitor) monitorResources() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			em.checkResourceLimits()
		}
	}
}

// checkResourceLimits checks if resource limits are exceeded.
func (em *ExecutionMonitor) checkResourceLimits() {
	// Check memory usage
	if em.security.MaxMemory > 0 {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		currentMem := memStats.Alloc - em.startMem

		if em.security.MaxMemory > 0 && int64(currentMem) > em.security.MaxMemory { //nolint:gosec // Safe conversion for memory limit check
			em.violations = append(em.violations,
				fmt.Sprintf("memory limit exceeded: %d bytes (limit: %d)",
					currentMem, em.security.MaxMemory))
			em.cancel()
			return
		}
	}

	// Check CPU time
	if em.security.MaxCPU > 0 {
		elapsed := time.Since(em.startTime)
		if elapsed > em.security.MaxCPU {
			em.violations = append(em.violations,
				fmt.Sprintf("CPU time limit exceeded: %v (limit: %v)",
					elapsed, em.security.MaxCPU))
			em.cancel()
			return
		}
	}
}

// Sandbox provides an isolated execution environment for plugins.
type Sandbox struct {
	security   *SecurityContext
	restricted map[string]bool
}

// NewSandbox creates a new sandbox environment.
func NewSandbox(security *SecurityContext) *Sandbox {
	return &Sandbox{
		security:   security,
		restricted: make(map[string]bool),
	}
}

// RestrictAPI marks an API as restricted in the sandbox.
func (sb *Sandbox) RestrictAPI(api string) {
	sb.restricted[api] = true
}

// CheckAPICall validates an API call against sandbox restrictions.
func (sb *Sandbox) CheckAPICall(module, function string) error {
	if !sb.security.Sandbox {
		return nil // Sandboxing disabled
	}

	// Check if API module is restricted
	if sb.restricted[module] {
		return fmt.Errorf("API %s is restricted in sandbox", module)
	}

	// Check against allowed APIs
	return sb.security.CheckAPIAccess(module)
}

// parseMemoryLimit parses memory limit strings like "100MB", "1GB".
func parseMemoryLimit(limit string) (int64, error) {
	if limit == "" {
		return 0, nil
	}

	// Simple parser for common memory units
	units := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range units {
		if len(limit) > len(suffix) && limit[len(limit)-len(suffix):] == suffix {
			numStr := limit[:len(limit)-len(suffix)]
			var num int64
			if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
				return 0, fmt.Errorf("invalid memory format: %s", limit)
			}
			return num * multiplier, nil
		}
	}

	return 0, fmt.Errorf("unsupported memory unit in: %s", limit)
}

// SecurityValidator validates plugin security configurations.
type SecurityValidator struct{}

// NewSecurityValidator creates a new security validator.
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{}
}

// ValidateManifest validates the security configuration in a plugin manifest.
func (sv *SecurityValidator) ValidateManifest(manifest *Manifest) error {
	// Validate security policy
	if manifest.Security.MaxMemory != "" {
		if _, err := parseMemoryLimit(manifest.Security.MaxMemory); err != nil {
			return fmt.Errorf("invalid memory limit in security policy: %w", err)
		}
	}

	if manifest.Security.MaxCPU != "" {
		if _, err := time.ParseDuration(manifest.Security.MaxCPU); err != nil {
			return fmt.Errorf("invalid CPU limit in security policy: %w", err)
		}
	}

	if manifest.Security.Timeout != "" {
		if _, err := time.ParseDuration(manifest.Security.Timeout); err != nil {
			return fmt.Errorf("invalid timeout in security policy: %w", err)
		}
	}

	// Validate capabilities against functions
	hasFunction := manifest.HasCapability(CapabilityFunction)
	if len(manifest.Functions) > 0 && !hasFunction {
		return fmt.Errorf("plugin defines functions but lacks 'function' capability")
	}

	// Validate dangerous capabilities
	dangerousCaps := []Capability{CapabilityProcess, CapabilityFileSystem, CapabilityNetwork}
	for _, cap := range dangerousCaps {
		if manifest.HasCapability(cap) && !manifest.Security.Sandbox {
			return fmt.Errorf("capability %s requires sandbox to be enabled", cap)
		}
	}

	return nil
}

// SecurityManager coordinates security enforcement across the plugin system.
type SecurityManager struct {
	validator *SecurityValidator
	policies  map[string]*SecurityContext
}

// NewSecurityManager creates a new security manager.
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		validator: NewSecurityValidator(),
		policies:  make(map[string]*SecurityContext),
	}
}

// RegisterPlugin registers a plugin's security policy.
func (sm *SecurityManager) RegisterPlugin(manifest *Manifest) error {
	if err := sm.validator.ValidateManifest(manifest); err != nil {
		return fmt.Errorf("security validation failed for plugin %s: %w", manifest.Name, err)
	}

	ctx, err := NewSecurityContext(manifest.Security)
	if err != nil {
		return fmt.Errorf("failed to create security context for plugin %s: %w", manifest.Name, err)
	}

	ctx.SetCapabilities(manifest.Capabilities)
	sm.policies[manifest.Name] = ctx

	return nil
}

// GetSecurityContext returns the security context for a plugin.
func (sm *SecurityManager) GetSecurityContext(pluginName string) (*SecurityContext, bool) {
	ctx, exists := sm.policies[pluginName]
	return ctx, exists
}

// CreateExecutionMonitor creates a new execution monitor for a plugin.
func (sm *SecurityManager) CreateExecutionMonitor(pluginName string) (*ExecutionMonitor, error) {
	ctx, exists := sm.policies[pluginName]
	if !exists {
		return nil, fmt.Errorf("no security policy found for plugin %s", pluginName)
	}

	return NewExecutionMonitor(ctx), nil
}
