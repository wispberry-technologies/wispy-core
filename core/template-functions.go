// Package core provides the main template engine and built-in tag functions for Wispy Core.
//
// The template engine supports:
//   - Per-render context (map[string]interface{} or custom structs)
//   - Custom tag functions (block and inline)
//   - Dot notation for nested map lookups (e.g., {{user.name}})
//   - Block tags (e.g., if, for, unless, case, assign, comment)
//
// Built-in tag functions:
//   - IfTemplate: Conditional rendering ({{if condition}}...{{end-if}})
//   - ForTemplate: Looping over slices/arrays ({{for item in items}}...{{end-for}})
//   - AssignTemplate: Assign a value to a variable in the context ({{assign var value}})
//   - CommentTemplate: Ignore content ({{comment}}...{{end-comment}})
//   - CaseTemplate: Switch/case logic ({{case var}}...{{when value}}...{{end-case}})
//   - UnlessTemplate: Render if condition is false ({{unless condition}}...{{end-unless}})
//   - DefineTemplate: Define a named block ({{define "name"}}...{{end-define}})
//   - RenderTemplate: Render a previously defined block ({{render "name"}})
//   - SanitizeTemplate: Escape HTML special characters for safe output ({{sanitize var}})
package core

import (
	"fmt"
	"strings"
	"wispy-core/models"
)

var IfTemplate = models.TemplateTag{
	Name: "if",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		var errs []error
		parts := strings.Fields(tagContents)
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("missing condition in if tag"))
			return pos, errs
		}
		cond := parts[1]
		endTag := "{{end-if}}"
		endIdx := strings.Index(raw[pos:], endTag)
		if endIdx == -1 {
			errs = append(errs, fmt.Errorf("could not find end tag for if"))
			return pos, errs
		}
		content := raw[pos : pos+endIdx]
		m, _ := ctx.Data.(map[string]interface{})
		if m != nil && m[cond] == true {
			sb.WriteString(content)
		}
		return pos + endIdx + len(endTag), errs
	},
}

var ForTemplate = models.TemplateTag{
	Name: "for",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		var errs []error
		parts := strings.Fields(tagContents)
		if len(parts) < 4 || parts[2] != "in" {
			errs = append(errs, fmt.Errorf("invalid for tag syntax"))
			return pos, errs
		}
		varName := parts[1]
		listName := parts[3]
		endTag := "{{end-for}}"
		endIdx := strings.Index(raw[pos:], endTag)
		if endIdx == -1 {
			errs = append(errs, fmt.Errorf("could not find end tag for for"))
			return pos, errs
		}
		content := raw[pos : pos+endIdx]
		m, _ := ctx.Data.(map[string]interface{})
		var list interface{}
		if m != nil {
			if strings.Contains(listName, ".") {
				parts := strings.Split(listName, ".")
				val := m[parts[0]]
				for _, p := range parts[1:] {
					if mm, ok := val.(map[string]interface{}); ok {
						val = mm[p]
					} else {
						val = nil
						break
					}
				}
				list = val
			} else {
				list = m[listName]
			}
		}
		if list != nil {
			engine := NewTemplateEngine(ctx.Engine.FuncMap)
			newCtx := models.TemplateContext{Data: m}
			switch v := list.(type) {
			case []string:
				for _, item := range v {
					localCtx := make(map[string]interface{})
					for k, val := range m {
						localCtx[k] = val
					}
					localCtx[varName] = item
					newCtx.Data = localCtx
					res, _ := Render(content, engine, &newCtx)
					sb.WriteString(res)
				}
			case []interface{}:
				for _, item := range v {
					localCtx := make(map[string]interface{})
					for k, val := range m {
						localCtx[k] = val
					}
					localCtx[varName] = item
					newCtx.Data = localCtx
					res, _ := Render(content, engine, &newCtx)
					sb.WriteString(res)
				}
			}
		}
		return pos + endIdx + len(endTag), errs
	},
}

// AssignTemplate assigns a value to a variable in the context for the rest of the render.
var AssignTemplate = models.TemplateTag{
	Name: "assign",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		errs := []error{}
		parts := strings.Fields(tagContents)
		if len(parts) < 3 {
			errs = append(errs, fmt.Errorf("assign tag requires variable name and value"))
			return pos, errs
		}
		varName := parts[1]
		value := strings.Join(parts[2:], " ")
		if m, ok := ctx.Data.(map[string]interface{}); ok {
			m[varName] = value
		}
		return pos, errs
	},
}

// CommentTemplate ignores all content between comment and end-comment.
var CommentTemplate = models.TemplateTag{
	Name: "comment",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		endTag := "{{end-comment}}"
		endIdx := strings.Index(raw[pos:], endTag)
		if endIdx == -1 {
			return pos, []error{fmt.Errorf("could not find end tag for comment")}
		}
		return pos + endIdx + len(endTag), nil
	},
}

// UnlessTemplate renders content only if the condition is false.
var UnlessTemplate = models.TemplateTag{
	Name: "unless",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		errs := []error{}
		parts := strings.Fields(tagContents)
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("missing condition in unless tag"))
			return pos, errs
		}
		cond := parts[1]
		endTag := "{{end-unless}}"
		endIdx := strings.Index(raw[pos:], endTag)
		if endIdx == -1 {
			errs = append(errs, fmt.Errorf("could not find end tag for unless"))
			return pos, errs
		}
		content := raw[pos : pos+endIdx]
		m, _ := ctx.Data.(map[string]interface{})
		if m != nil && m[cond] != true {
			sb.WriteString(content)
		}
		return pos + endIdx + len(endTag), errs
	},
}

// CaseTemplate implements a simple switch/case logic.
var CaseTemplate = models.TemplateTag{
	Name: "case",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		errs := []error{}
		parts := strings.Fields(tagContents)
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("missing variable in case tag"))
			return pos, errs
		}
		varName := parts[1]
		endTag := "{{end-case}}"
		whenTag := "{{when "
		endIdx := strings.Index(raw[pos:], endTag)
		if endIdx == -1 {
			errs = append(errs, fmt.Errorf("could not find end tag for case"))
			return pos, errs
		}
		block := raw[pos : pos+endIdx]
		m, _ := ctx.Data.(map[string]interface{})
		var val interface{}
		if m != nil {
			val = m[varName]
		}
		// Find all when blocks
		searchPos := 0
		for searchPos < len(block) {
			wStart := strings.Index(block[searchPos:], whenTag)
			if wStart == -1 {
				break
			}
			wStart += searchPos
			wEnd := strings.Index(block[wStart:], "}}")
			if wEnd == -1 {
				break
			}
			wEnd += wStart
			whenVal := strings.TrimSpace(block[wStart+len(whenTag) : wEnd])
			// Find next when or end-case
			nextWhen := strings.Index(block[wEnd:], whenTag)
			endBlock := len(block)
			if nextWhen != -1 {
				endBlock = wEnd + nextWhen
			}
			if fmt.Sprint(val) == whenVal {
				sb.WriteString(block[wEnd+2 : endBlock])
				break
			}
			searchPos = endBlock
		}
		return pos + endIdx + len(endTag), errs
	},
}

// DefineTemplate allows defining a named block: {{define "name"}}...{{end-define}}
var DefineTemplate = models.TemplateTag{
	Name: "define",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		errs := []error{}
		parts := strings.Fields(tagContents)
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("define tag requires a block name"))
			return pos, errs
		}
		name := strings.Trim(parts[1], "\"'")
		endTag := ctx.Engine.StartTag + "end-define" + ctx.Engine.EndTag
		endIdx := strings.Index(raw[pos:], endTag)
		if endIdx == -1 {
			errs = append(errs, fmt.Errorf("could not find end tag for define: %s", name))
			return pos, errs
		}
		content := raw[pos : pos+endIdx]
		// Store block in InternalContext (as map[string]string)
		var blockMap map[string]string
		if ctx.InternalContext == nil {
			blockMap = map[string]string{}
			ctx.InternalContext = blockMap
		} else if bm, ok := ctx.InternalContext.(map[string]string); ok {
			blockMap = bm
		} else {
			blockMap = map[string]string{}
			ctx.InternalContext = blockMap
		}
		blockMap[name] = content
		return pos + endIdx + len(endTag), errs
	},
}

// RenderTemplate renders a previously defined block: {{render "name"}}
var RenderTemplate = models.TemplateTag{
	Name: "render",
	Render: func(ctx TemplateCtx, sb *strings.Builder, tagContents, raw string, pos int) (int, []error) {
		errs := []error{}
		parts := strings.Fields(tagContents)
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("render tag requires a block name"))
			return pos, errs
		}
		name := strings.Trim(parts[1], "\"'")
		var blockMap map[string]string
		if ctx.InternalContext != nil {
			if bm, ok := ctx.InternalContext.(map[string]string); ok {
				blockMap = bm
			}
		}
		if blockMap == nil {
			errs = append(errs, fmt.Errorf("no blocks defined for render: %s", name))
			return pos, errs
		}
		block, ok := blockMap[name]
		if !ok {
			errs = append(errs, fmt.Errorf("block not found for render: %s", name))
			return pos, errs
		}
		// Render the block content with the current context
		engine := NewTemplateEngine(ctx.Engine.FuncMap)
		res, _ := Render(block, engine, ctx)
		sb.WriteString(res)
		return pos, errs
	},
}

// DefaultFunctionMap provides all built-in tags.
var DefaultFunctionMap = models.FunctionMap{
	"if":      IfTemplate,
	"for":     ForTemplate,
	"assign":  AssignTemplate,
	"comment": CommentTemplate,
	"unless":  UnlessTemplate,
	"case":    CaseTemplate,
	"define":  DefineTemplate,
	"render":  RenderTemplate,
}
