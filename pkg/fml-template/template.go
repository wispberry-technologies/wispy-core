package template

import (
	"wispy-core/pkg/common"

	"github.com/microcosm-cc/bluemonday"
)

type EngineConfig struct {
	TemplateTags      []*TemplateTag
	TemplateFuncs     []*TemplateFunc
	Sanitizer         EngineSanitizer
	StartDelim        string                  // Delimiter for template tags, Default "{{"
	EndDelim          string                  // Delimiter for template tags, Default "}}"
	DataAdapters      map[string]*DataAdapter // List of data adapters for resolving data in templates keyed by data prefix
	GlobalDataContext map[string]any          // Default data to be used in templates, can be overridden by DataAdapters
}

type Engine struct {
	TemplateTags      map[string]TemplateTagResolver
	templateTagNames  map[string]bool
	TemplateFuncs     map[string]TemplateFuncResolver
	templateFuncNames map[string]bool
	Sanitizer         EngineSanitizer
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

var UGCPolicy *bluemonday.Policy

func init() {
	UGCPolicy = bluemonday.UGCPolicy()
}

func NewEngine(config EngineConfig) *Engine {
	engine := &Engine{
		Sanitizer:         EngineSanitizer{},
		TemplateTags:      make(map[string]TemplateTagResolver, 6),
		templateTagNames:  make(map[string]bool, 6),
		TemplateFuncs:     make(map[string]TemplateFuncResolver, 10),
		templateFuncNames: make(map[string]bool, 10),
		StartDelim:        "{{",
		EndDelim:          "}}",
		Variables:         make(map[string]any, 10),
		Blocks:            make(map[string]any, 10),
		DataAdapters:      make(map[string]*DataAdapter, 5),
		GlobalDataContext: make(map[string]any, 10),
		LocalDataContext:  make(map[string]any, 10),
		Flags: map[string]bool{
			"else": false, // Flag to track the previous tag returned false & suppress the else branch
		},
		// Used to store metadata about the rendering process until the template is rendered
		// This can be used to store information like execution time, assets used, associated data, etc.
		Metadata: make(map[string]any, 2),
	}

	// Initialize the sanitizer if no custom function is provided
	if config.Sanitizer.SanitizeFunc != nil {
		if config.Sanitizer.Enabled {
			engine.Sanitizer.SanitizeFunc = config.Sanitizer.SanitizeFunc
			engine.Sanitizer = EngineSanitizer{
				SanitizeFunc:   func(input string) string { return UGCPolicy.Sanitize(input) },
				SanitizePolicy: UGCPolicy,
			}
		} else {
			common.Warning("Sanitization is disabled, using no-op sanitizer function")
			engine.Sanitizer.SanitizePolicy = nil
			engine.Sanitizer.SanitizeFunc = func(input string) string {
				// If sanitization is disabled, return the input unchanged
				return input
			}
		}
	}

	// Check for non-nil config options and populate
	switch {
	case config.TemplateFuncs != nil:
		for _, fn := range config.TemplateFuncs {
			engine.templateFuncNames[fn.Name] = true
			engine.TemplateFuncs[fn.Name] = fn.Handler
		}
	case config.TemplateTags != nil:
		for _, tag := range config.TemplateTags {
			engine.templateTagNames[tag.Name] = true
			engine.TemplateTags[tag.Name] = tag.Handler
		}
	case config.StartDelim != "":
		engine.StartDelim = config.StartDelim
	case config.EndDelim != "":
		engine.EndDelim = config.EndDelim
	}

	return engine
}
