package main

import (
	"fmt"
	"wispy-core/core"
)

func errorResilienceDemo() {
	// Create a template engine
	engine := core.NewTemplateEngine(core.DefaultFunctionMap())

	// Create context with some data
	ctx := core.NewTemplateContext(map[string]interface{}{
		"user": map[string]interface{}{
			"name":   "John Doe",
			"active": true,
		},
		"site": map[string]interface{}{
			"title": "My Website",
		},
	}, engine)

	// Template with various error conditions
	template := `
<h1>{{ site.title }}</h1>
<p>Welcome {{ user.name }}!</p>
<p>Status: {{ user.status }}</p>  <!-- unresolved variable -->
<p>Email: {{ user.email | default:"No email" }}</p>  <!-- unresolved variable with filter -->
<p>Unknown filter: {{ user.name | unknown_filter }}</p>  <!-- unknown filter -->

{% if user.active %}
<p>User is active</p>
{% endif %}

{% if user.missing_field %}
<p>This won't show because missing_field is nil</p>
{% endif %}

{% if user.name | bad_filter %}
<p>This might not work due to bad filter</p>
{% endif %}

{% unknown_tag %}

<p>This content should still render despite errors above</p>

{% for item in user.hobbies %}
<li>{{ item }}</li>
{% endfor %}

<p>Final content: {{ site.title }} is working!</p>
`

	// Render the template
	result, errors := engine.Render(template, ctx)

	fmt.Println("=== Rendered Output ===")
	fmt.Println(result)
	fmt.Println("\n=== Errors (but rendering continued) ===")
	for i, err := range errors {
		fmt.Printf("%d. %v\n", i+1, err)
	}
	fmt.Printf("\nTotal errors: %d\n", len(errors))
	fmt.Println("\nAs you can see, even with errors, the template rendered successfully!")
}
