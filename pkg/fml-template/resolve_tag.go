package template

import (
	"fmt"
	"strings"
	"wispy-core/pkg/common"
)

func (e *Engine) ResolveRawTag(raw string, sb *strings.Builder, startPos, endPos int) (newPos int, errs []error) {
	var content = raw[startPos+len(e.StartDelim) : endPos]
	tag, args := common.SplitTagParts(content)

	if tag == "" {
		errs = append(errs, fmt.Errorf("empty tag found at position %d", startPos))
		return endPos + len(e.EndDelim), errs
	}

	var (
		value any
		err   error
	)
	switch {
	case tag[0] == '.':
		var ok bool
		value, ok = e.GetValue(tag)
		if !ok {
			errs = append(errs, fmt.Errorf("unknown variable '%s' at position %d", tag[1:], startPos))
			return endPos + len(e.EndDelim), errs
		}
	case tag[0] == '$':
		if len(args) >= 3 {
			if args[1] == ":=" {
				value, err = e.ResolveArguments(nil, args[2:])
				if err != nil {
					errs = append(errs, err)
					return endPos + len(e.EndDelim), errs
				}
				//
				e.SetVariable(tag[1:], value)
				return endPos + len(e.EndDelim), errs
			}
		} else {
			// If no arguments are provided, just get the variable value
			value, _ = e.GetVariable(tag[1:])
		}
	case e.templateTagNames[tag]:
		tagHandler := e.TemplateTags[tag]
		newPos, tagErrs := tagHandler(e, sb, args, raw, endPos+len(e.EndDelim))
		if len(tagErrs) > 0 {
			errs = append(errs, tagErrs...)
		}
		return newPos, errs

	// Edge cases for widow end tags
	case strings.HasPrefix(tag, "end"):
		errs = append(errs, fmt.Errorf("unexpected 'end' tag '%s' at position %d", tag, startPos))
		return endPos + len(e.EndDelim), errs
	default:
		errs = append(errs, fmt.Errorf("unknown tag '%s' at position %d", tag, startPos))
		return endPos + len(e.EndDelim), errs
	}

	// If additional functions pipe values
	resolved, err := e.ResolveArguments(value, args)
	if err != nil {
		errs = append(errs, err)
		return endPos + len(e.EndDelim), errs
	} else {
		if resolved != nil {
			sb.WriteString(fmt.Sprintf("%v", resolved))
		}
	}

	return endPos + len(e.EndDelim), errs
}

func (e *Engine) ResolveArguments(value any, args []string) (any, error) {
	if len(args) == 0 {
		return value, nil
	}

	// If we have a single argument and no initial value, resolve that argument
	if value == nil && len(args) == 1 {
		arg := args[0]
		if strings.HasPrefix(arg, ".") {
			// This is a data reference like .User.Name
			// First try to get it from the local context
			if arg == "." {
				// Just the dot means "current item"
				return e.GetLocalDataContext()["."], nil
			}

			val, ok := e.GetLocalDataContextValue(arg[1:])
			if ok {
				return val, nil
			}

			// Try resolving as a path
			val, ok = e.GetValue(arg)
			if !ok {
				return nil, fmt.Errorf("unknown variable '%s'", arg)
			}
			return val, nil
		} else if strings.HasPrefix(arg, "$") {
			// This is a variable reference like $name
			val, ok := e.GetVariable(arg[1:])
			if !ok {
				return nil, fmt.Errorf("unknown variable '%s'", arg[1:])
			}
			return val, nil
		} else {
			// Try to interpret as a literal value
			return arg, nil
		}
	}

	return value, nil
}
