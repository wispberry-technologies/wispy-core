package site

import (
	"bytes"
	"html/template"
	"io"
	"wispy-core/common"
	"wispy-core/core"
	"wispy-core/tpl"
)

type siteTplEngine struct {
	tplEngine     tpl.Engine
	site          core.Site
	wispyTailTrie *common.Trie
}

func (ste *siteTplEngine) LoadTemplate(templatePath string) (*template.Template, error) {
	return ste.tplEngine.LoadTemplate(templatePath)
}

func (ste *siteTplEngine) RenderTemplate(templatePath string, data tpl.TemplateData) (string, error) {
	tmpl, err := ste.LoadTemplate(templatePath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (ste *siteTplEngine) RenderTemplateTo(w io.Writer, templatePath string, data tpl.TemplateData) error {
	tmpl, err := ste.LoadTemplate(templatePath)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

func (ste *siteTplEngine) ScanPages() ([]string, error) {
	pages, err := ste.tplEngine.ScanPages()
	if err != nil {
		return nil, err
	}
	return pages, nil
}

func (ste *siteTplEngine) GetTrie() *common.Trie {
	if ste.wispyTailTrie == nil {
		ste.wispyTailTrie = common.NewTrie()
	}
	return ste.wispyTailTrie
}
