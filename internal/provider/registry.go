package provider

import (
	"fmt"
	"strings"
	"sync"
)

// Registry manages all registered providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	models    map[string]string // model_id -> provider_name
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		models:    make(map[string]string),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(p Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = p

	// Index models
	for _, model := range p.SupportedModels() {
		r.models[model.ID] = name
	}

	return nil
}

// GetByName returns a provider by name
func (r *Registry) GetByName(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return p, nil
}

// GetByModel automatically determines the provider by model
func (r *Registry) GetByModel(modelID string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Parse "provider/model" format
	if strings.Contains(modelID, "/") {
		parts := strings.SplitN(modelID, "/", 2)
		providerName := parts[0]
		if p, ok := r.providers[providerName]; ok {
			return p, nil
		}
	}

	// Search by model index
	if providerName, ok := r.models[modelID]; ok {
		return r.providers[providerName], nil
	}

	return nil, fmt.Errorf("no provider found for model %s", modelID)
}

// ListProviders returns a list of all providers
func (r *Registry) ListProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// ListModels returns a list of all models
func (r *Registry) ListModels() []Model {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []Model
	for _, p := range r.providers {
		models = append(models, p.SupportedModels()...)
	}
	return models
}
