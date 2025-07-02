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

type renderState struct {
	mu        sync.Mutex
	title     string
	inlineCSS string
	inlineJS  string
	styles    []StyleAsset
	scripts   []ScriptAsset
	body      string
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
	AddHeadInlineCSS(css string)
	AddHeadInlineJS(js string)
	SetBody(content string)
	GetBody() string
}

func NewRenderState() RenderState {
	return &renderState{
		title:     "",
		inlineCSS: "",
		inlineJS:  "",
		styles:    []StyleAsset{},
		scripts:   []ScriptAsset{},
		body:      "",
	}
}

func (rs *renderState) GetHeadTitle() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.title
}

func (rs *renderState) GetHeadInlineCSS() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.inlineCSS
}

func (rs *renderState) GetHeadInlineJS() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.inlineJS
}

func (rs *renderState) GetHeadStyles() []StyleAsset {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.styles
}

func (rs *renderState) GetHeadScripts() []ScriptAsset {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.scripts
}

func (rs *renderState) SetHeadTitle(title string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.title = title
}

func (rs *renderState) AddHeadInlineCSS(css string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.inlineCSS += css
}

func (rs *renderState) AddHeadInlineJS(js string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.inlineJS += js
}

func (rs *renderState) AddStyles(styles StyleAsset) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.styles = append(rs.styles, styles)
}

func (rs *renderState) AddScripts(scripts ScriptAsset) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.scripts = append(rs.scripts, scripts)
}

func (rs *renderState) SetBody(content string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.body = content
}

func (rs *renderState) GetBody() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.body
}
