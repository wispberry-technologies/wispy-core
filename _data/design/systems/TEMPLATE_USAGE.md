# Template Usage Guide

## Important Note
Do not include HTML comments with template syntax (like `{{template ...}}`) in template files, as this causes recursion issues with Go's html/template package.

## Atoms

### Button
```
{{template "atoms/button" dict "text" "Sign in" "type" "submit" "style" "btn-primary" "size" "btn-md" "class" "w-full"}}
```

### Input
```
{{template "atoms/input" dict "type" "email" "name" "email" "placeholder" "Email address" "value" .email "required" true "class" "input-bordered"}}
```

### Label
```
{{template "atoms/label" dict "for" "email" "text" "Email address" "class" "label-text"}}
```

### Alert
```
{{template "atoms/alert" dict "type" "alert-error" "message" "Login failed" "icon" true}}
```

### Badge
```
{{template "atoms/badge" dict "text" "New" "style" "badge-primary" "size" "badge-sm"}}
```

### Card
```
{{template "atoms/card" dict "title" "Forms" "description" "Manage form submissions" "link" "/wispy-cms/forms" "linkText" "View Forms"}}
```

## Components

### CMS Navbar
```
{{template "components/cms-navbar" dict "currentPage" "dashboard"}}
```

### Form Field
```
{{template "components/form-field" dict "label" "Email" "type" "email" "name" "email" "placeholder" "Enter your email" "value" .email "required" true "error" .emailError}}
```

### Login Form
```
{{template "components/login-form" .}}
```

### Stats Dashboard
```
{{template "components/stats-dashboard" dict "formCount" 5 "submissionCount" 23 "settingsCount" 8}}
```

### Breadcrumb
```
{{template "components/breadcrumb" dict "items" (slice (dict "text" "Home" "href" "/") (dict "text" "Forms" "href" "/forms") (dict "text" "Submissions"))}}
```

### Modal
```
{{template "components/modal" dict "id" "my-modal" "title" "Confirm Action" "content" "Are you sure?" "buttons" (slice (dict "text" "Cancel" "style" "btn-outline") (dict "text" "Confirm" "style" "btn-primary"))}}
```

### Loading
```
{{template "components/loading" dict "size" "loading-lg" "text" "Loading..."}}
```

### Empty State
```
{{template "components/empty-state" dict "title" "No submissions yet" "description" "When users submit forms, they'll appear here." "action" (dict "text" "Create Form" "href" "/forms/new")}}
```

### Table
```
{{template "components/table" dict "headers" (slice "Name" "Email" "Date") "rows" .submissions "actions" true}}
```
