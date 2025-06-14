package core

import (
	"testing"
)

func BenchmarkTemplateEngine_SimpleVariable(b *testing.B) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"name": "World",
	}, engine)
	template := "Hello {{ name }}!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Render(template, ctx)
	}
}

func BenchmarkTemplateEngine_ComplexTemplate(b *testing.B) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"title": "My Blog",
		"posts": []interface{}{
			map[string]interface{}{
				"title":   "First Post",
				"content": "This is the first post",
				"author":  "John",
			},
			map[string]interface{}{
				"title":   "Second Post",
				"content": "This is the second post",
				"author":  "Jane",
			},
		},
		"showPosts": true,
	}, engine)

	template := `<h1>{{ title | upcase }}</h1>
{% if showPosts %}
<div class="posts">
{% for post in posts %}
  <article>
    <h2>{{ post.title }}</h2>
    <p>{{ post.content }}</p>
    <small>By {{ post.author }}</small>
  </article>
{% endfor %}
</div>
{% endif %}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Render(template, ctx)
	}
}

func BenchmarkTemplateEngine_NestedLoops(b *testing.B) {
	engine := NewTemplateEngine(DefaultFunctionMap())

	// Create nested data structure
	categories := []interface{}{}
	for i := 0; i < 5; i++ {
		items := []interface{}{}
		for j := 0; j < 10; j++ {
			items = append(items, map[string]interface{}{
				"name":  "Item " + string(rune('A'+j)),
				"value": j * 10,
			})
		}
		categories = append(categories, map[string]interface{}{
			"name":  "Category " + string(rune('1'+i)),
			"items": items,
		})
	}

	ctx := NewTemplateContext(map[string]interface{}{
		"categories": categories,
	}, engine)

	template := `{% for category in categories %}
<div class="category">
  <h3>{{ category.name }}</h3>
  <ul>
  {% for item in category.items %}
    <li>{{ item.name }}: {{ item.value }}</li>
  {% endfor %}
  </ul>
</div>
{% endfor %}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Render(template, ctx)
	}
}
