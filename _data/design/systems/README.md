# Wispy CMS Design System

This document describes the atomic design system implemented for Wispy CMS, following the principles of atomic design and built with daisyUI v5.

## Architecture

The design system is organized into three main layers:

- **Atoms** (`/atoms/`): Basic building blocks (buttons, inputs, labels, alerts, etc.)
- **Components** (`/components/`): Composed reusable components (login forms, navigation, tables, etc.)
- **Templates** (`/templates/cms/`): Page-level templates that combine components

## Using the Design System

### Atoms

Atoms are the basic building blocks of the UI. They're designed to be flexible and reusable.

#### Button Atom
```html
{{template "atoms/button" dict "text" "Sign in" "type" "submit" "style" "btn-primary" "size" "btn-md" "class" "w-full"}}
```

Parameters:
- `text`: Button text (required)
- `type`: Button type (default: "button")
- `style`: daisyUI button style (default: "btn-primary")
- `size`: Button size (default: "btn-md")
- `class`: Additional CSS classes

#### Input Atom
```html
{{template "atoms/input" dict "type" "email" "name" "email" "placeholder" "Email address" "value" .email "required" true "class" "input-bordered"}}
```

Parameters:
- `type`: Input type (default: "text")
- `name`: Input name (required)
- `id`: Input ID (defaults to name)
- `placeholder`: Placeholder text
- `value`: Input value
- `required`: Whether field is required (boolean)
- `autocomplete`: Autocomplete attribute
- `class`: Additional CSS classes

#### Alert Atom
```html
{{template "atoms/alert" dict "type" "alert-error" "message" "Login failed" "icon" true}}
```

Parameters:
- `type`: Alert type (default: "alert-info")
- `message`: Alert message (required)
- `icon`: Show icon (boolean)

#### Card Atom
```html
{{template "atoms/card" dict "title" "Forms" "description" "Manage form submissions" "link" "/wispy-cms/forms" "linkText" "View Forms"}}
```

Parameters:
- `title`: Card title
- `description`: Card description
- `link`: Link URL
- `linkText`: Link text (default: "Learn More")

#### Badge Atom
```html
{{template "atoms/badge" dict "text" "New" "style" "badge-primary" "size" "badge-sm"}}
```

Parameters:
- `text`: Badge text (required)
- `style`: Badge style (default: "badge-neutral")
- `size`: Badge size (default: "badge-md")

### Components

Components are composed of atoms and provide more complex functionality.

#### Login Form Component
```html
{{template "components/login-form" .}}
```

Expects context with:
- `.hasError`: Boolean indicating if there's an error
- `.errorMessage`: Error message to display
- `.email`: Email value to preserve on error

#### CMS Navigation Component
```html
{{template "components/cms-navbar" dict "currentPage" "dashboard"}}
```

Parameters:
- `currentPage`: Current page identifier for highlighting active nav item

#### Stats Dashboard Component
```html
{{template "components/stats-dashboard" dict "formCount" 5 "submissionCount" 23 "settingsCount" 8}}
```

Parameters:
- `formCount`: Number of forms
- `submissionCount`: Number of submissions
- `settingsCount`: Number of settings

#### Form Field Component
```html
{{template "components/form-field" dict "label" "Email" "type" "email" "name" "email" "placeholder" "Enter your email" "value" .email "required" true "error" .emailError}}
```

Parameters:
- `label`: Field label
- `type`: Input type (supports "textarea", "select", and standard input types)
- `name`: Field name (required)
- `placeholder`: Placeholder text
- `value`: Field value
- `required`: Whether field is required (boolean)
- `error`: Error message to display
- `help`: Help text
- `options`: For select fields, array of {value, label} objects
- `rows`: For textarea fields, number of rows

#### Modal Component
```html
{{template "components/modal" dict "id" "my-modal" "title" "Confirm Action" "content" "Are you sure?" "buttons" (slice (dict "text" "Cancel" "style" "btn-outline") (dict "text" "Confirm" "style" "btn-primary"))}}
```

Parameters:
- `id`: Modal ID (required)
- `title`: Modal title
- `content`: Modal content text
- `customContent`: Custom HTML content
- `buttons`: Array of button objects with `text`, `style`, and optional `onclick`

#### Breadcrumb Component
```html
{{template "components/breadcrumb" dict "items" (slice (dict "text" "Home" "href" "/") (dict "text" "Forms" "href" "/forms") (dict "text" "Submissions"))}}
```

Parameters:
- `items`: Array of breadcrumb items with `text` and optional `href`

#### Empty State Component
```html
{{template "components/empty-state" dict "title" "No submissions yet" "description" "When users submit forms, they'll appear here." "action" (dict "text" "Create Form" "href" "/forms/new")}}
```

Parameters:
- `title`: Empty state title
- `description`: Empty state description
- `action`: Optional action object with `text` and `href`

## Theme System

The design system uses daisyUI v5 themes. Current themes available:

- `robot-green` (default)
- `midnight-wisp`
- `pale-wisp`

Themes can be changed by updating the `data-theme` attribute on the `<html>` tag.

## CSS Variables

The system uses CSS custom properties for consistent theming:

- `oklch(var(--b1))`: Base background color
- `oklch(var(--bc))`: Base content color
- `oklch(var(--p))`: Primary color
- `oklch(var(--s))`: Secondary color
- `oklch(var(--a))`: Accent color

## Best Practices

1. **Use atoms for basic elements**: Always use atoms for buttons, inputs, and other basic elements
2. **Compose components from atoms**: Build complex components by combining atoms
3. **Keep components focused**: Each component should have a single responsibility
4. **Use consistent naming**: Follow the established naming conventions
5. **Provide defaults**: Always provide sensible defaults for optional parameters
6. **Document parameters**: Clearly document all parameters and their expected values

## File Structure

```
/_data/design/systems/
├── atoms/
│   ├── button.html
│   ├── input.html
│   ├── label.html
│   ├── alert.html
│   ├── card.html
│   └── badge.html
├── components/
│   ├── login-form.html
│   ├── cms-navbar.html
│   ├── stats-dashboard.html
│   ├── form-field.html
│   ├── modal.html
│   ├── breadcrumb.html
│   ├── loading.html
│   └── empty-state.html
└── themes/
    ├── robot-green.css
    ├── midnight-wisp.css
    └── pale-wisp.css
```

## Contributing

When adding new components:

1. Start with atoms if you need new basic elements
2. Build components by combining atoms
3. Follow the established naming conventions
4. Document all parameters and provide examples
5. Test with all available themes
6. Update this README with new components
