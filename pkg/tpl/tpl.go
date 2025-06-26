package tpl

import (
	textTemplate "wispy-core/pkg/go_templates/textTemplate"
)

type EngineConfig struct {
	TemplateFuncs     textTemplate.FuncMap
	StartDelim        string                  // Delimiter for template tags, Default "{{"
	EndDelim          string                  // Delimiter for template tags, Default "}}"
	DataAdapters      map[string]*DataAdapter // List of data adapters for resolving data in templates keyed by data prefix
	GlobalDataContext map[string]any          // Default data to be used in templates, can be overridden by DataAdapters
}

type Engine struct {
	TemplateFuncs     textTemplate.FuncMap
	templateFuncNames map[string]bool
	StartDelim        string
	EndDelim          string
	Variables         map[string]any          // For storing variables set in templates
	Blocks            map[string]any          // For storing block content
	DataAdapters      map[string]*DataAdapter // For resolving data in templates keyed by data prefix
	GlobalDataContext map[string]any          // Default data to be used in templates, can be overridden by DataAdapters
	LocalDataContext  map[string]any          // Local context for template rendering, can be used to override or extend data
	Flags             map[string]bool         // Flags for controlling template behavior such as flagging else branches
	Metadata          map[string]any          // Metadata about the rendering process, e.g., execution time, assets used, etc.
}

type DataAdapter struct {
	Prefix   string                                                   // Prefix for the data adapter, e.g., "user."
	Writable bool                                                     // Whether the data adapter allows writing
	GetData  func(e *Engine, pathKeys ...string) (value any, ok bool) // Function to get data by path keys
	SetData  func(e *Engine, value any, pathKeys ...string)           // Function to set data by path keys
}

func NewEngine(config EngineConfig) *Engine {
	engine := &Engine{
		TemplateFuncs:     make(textTemplate.FuncMap, 10),
		templateFuncNames: make(map[string]bool, 10),
		StartDelim:        "{{",
		EndDelim:          "}}",
		// Variables:         make(map[string]any, 10),
		// Blocks:            make(map[string]any, 10),
		// DataAdapters:      make(map[string]*DataAdapter, 5),
		// GlobalDataContext: make(map[string]any, 10),
		// LocalDataContext:  make(map[string]any, 10),
		// Used to store metadata about the rendering process until the template is rendered
		// This can be used to store information like execution time, assets used, associated data, etc.
		Metadata: make(map[string]any, 2),
	}

	// Check for non-nil config options and populate
	switch {
	case config.TemplateFuncs != nil:
		for name, fn := range config.TemplateFuncs {
			engine.TemplateFuncs[name] = fn
		}
	case config.StartDelim != "":
		engine.StartDelim = config.StartDelim
	case config.EndDelim != "":
		engine.EndDelim = config.EndDelim
	}

	return engine
}

// /*
// //	Engine CRUD Operations
// */
// // Add & Get blocks for template content
// func (e *Engine) AddBlock(name string, content any) {
// 	if e.Blocks == nil {
// 		e.Blocks = make(map[string]any)
// 	}
// 	e.Blocks[name] = content
// }

// func (e *Engine) GetBlock(name string) (any, bool) {
// 	if e.Blocks == nil {
// 		return nil, false
// 	}
// 	content, exists := e.Blocks[name]
// 	return content, exists
// }

// // Get & Set local context variables
// func (e *Engine) SetVariable(name string, value any) {
// 	if e.Variables == nil {
// 		e.Variables = make(map[string]any)
// 	}
// 	e.Variables[name] = value
// }

// func (e *Engine) GetVariable(name string) (any, bool) {
// 	if e.Variables == nil {
// 		return nil, false
// 	}
// 	value, exists := e.Variables[name]
// 	return value, exists
// }

// // GetValue retrieves a value from the template context.
// func (e *Engine) GetValue(key string, pathKeys ...string) (any, bool) {
// 	if key == "." {
// 		// If the key is just a dot, return the current context
// 		common.Debug("Returning local data context for key: %s", key)
// 		return e.GetLocalDataContext(), true
// 	}

// 	key = strings.TrimPrefix(key, ".") // Remove leading dot if present
// 	common.Debug("Resolving value for key: %s with path keys: %v", key, pathKeys)
// 	value, ok := e.ResolveValue(key, pathKeys...)
// 	return value, ok
// }

// // LocalDataContext allows setting a local context for the template rendering.
// func (e *Engine) SetLocalDataContext(ctx map[string]any) {
// 	if e.LocalDataContext == nil {
// 		e.LocalDataContext = make(map[string]any)
// 	}
// 	for k, v := range ctx {
// 		e.LocalDataContext[k] = v
// 	}
// }

// // SetLocalDataContextValue sets a value in the local data context for the template.
// func (e *Engine) SetLocalDataContextValue(key string, value any) {
// 	if e.LocalDataContext == nil {
// 		e.LocalDataContext = make(map[string]any)
// 	}
// 	e.LocalDataContext[key] = value
// }

// // GetLocalDataContextValue retrieves a value from the local data context for the template.
// func (e *Engine) GetLocalDataContextValue(key string) (any, bool) {
// 	if e.LocalDataContext == nil {
// 		return nil, false
// 	}
// 	value, exists := e.LocalDataContext[key]
// 	return value, exists
// }

// // GetLocalDataContext retrieves the local data context for the template.
// func (e *Engine) GetLocalDataContext() map[string]any {
// 	if e.LocalDataContext == nil {
// 		return make(map[string]any)
// 	}
// 	return e.LocalDataContext
// }

// // Getters
// func (e *Engine) GetDataAdapter(name string) (*DataAdapter, bool) {
// 	if e.DataAdapters == nil {
// 		return nil, false
// 	}
// 	adapter, exists := e.DataAdapters[name]
// 	return adapter, exists
// }
