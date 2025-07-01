package tpl

import (
	"sync"
)

type StyleAsset struct {
	Src      string
	Location string
	Priority int
	Attrs    map[string]string
	Type     int
}

type ScriptAsset struct {
	Src   string
	Async bool
	Defer bool
}

type renderStateHeader struct {
	title     string
	inlineCSS string
	inlineJS  string
	styles    []StyleAsset
	scripts   []ScriptAsset
}

type renderState struct {
	mu   sync.Mutex
	head renderStateHeader
	body string
}

type RenderState interface {
	// Defer(key string, fn func() error)
	GetHeadTitle() string
	GetHeadInlineCSS() string
	GetHeadInlineJS() string
	GetHeadStyles() []StyleAsset
	AddStyles(styles StyleAsset)
	GetHeadScripts() []ScriptAsset
	AddScripts(scripts ScriptAsset)
	SetHeadTitle(title string)
	SetHeadInlineCSS(css string)
	SetHeadInlineJS(js string)
	SetBody(content string)
	Body() string
	// ExecuteDeferred() error
}

func NewRenderState() RenderState {
	return &renderState{
		head: renderStateHeader{
			title:     "",
			inlineCSS: "",
			inlineJS:  "",
			styles:    []StyleAsset{},
			scripts:   []ScriptAsset{},
		},
		body: "",
	}
}

func (rs *renderState) GetHeadTitle() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.head.title
}

func (rs *renderState) GetHeadInlineCSS() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.head.inlineCSS
}

func (rs *renderState) GetHeadInlineJS() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.head.inlineJS
}

func (rs *renderState) GetHeadStyles() []StyleAsset {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.head.styles
}

func (rs *renderState) GetHeadScripts() []ScriptAsset {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.head.scripts
}

// Push operations (thread-safe)
// func (rs *renderState) Defer(key string, fn func() error) {
// 	rs.mu.Lock()
// 	defer rs.mu.Unlock()
// 	rs.deferred = append(rs.deferred, &deferredTask{Key: key, Exec: fn})
// }

func (rs *renderState) SetHeadTitle(title string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.head.title = title
}

func (rs *renderState) SetHeadInlineCSS(css string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.head.inlineCSS = css
}

func (rs *renderState) SetHeadInlineJS(js string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.head.inlineJS = js
}

func (rs *renderState) AddStyles(styles StyleAsset) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.head.styles = append(rs.head.styles, styles)
}

func (rs *renderState) AddScripts(scripts ScriptAsset) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.head.scripts = append(rs.head.scripts, scripts)
}

func (rs *renderState) SetBody(content string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.body = content
}

func (rs *renderState) Body() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.body
}

// func (rs *renderState) ExecuteDeferred() error {
// 	rs.mu.Lock()
// 	defer rs.mu.Unlock()

// 	for _, task := range rs.deferred {
// 		if err := task.Exec(); err != nil {
// 			return err
// 		}
// 	}
// 	rs.deferred = nil // Clear after execution
// 	return nil
// }
