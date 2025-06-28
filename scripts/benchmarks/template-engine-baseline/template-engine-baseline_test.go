package main

import (
	"io"
	"testing"

	html_template "html/template"
	text_template "text/template"
)

// Benchmark data structure
type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

type Job struct {
	Title    string
	Company  string
	Location Address
	Years    int
}

type Person struct {
	Name      string
	Age       int
	Email     string
	Friends   []string
	Addresses []Address
	Jobs      []Job
	Metadata  map[string]string
	Active    bool
	Scores    []int
	Tags      []string
}

var testPerson = Person{
	Name:    "John Doe",
	Age:     30,
	Email:   "john@example.com",
	Friends: []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Judy"},
	Addresses: []Address{
		{Street: "123 Main St", City: "Metropolis", State: "NY", Zip: "10001"},
		{Street: "456 Elm St", City: "Gotham", State: "NJ", Zip: "07001"},
		{Street: "789 Oak St", City: "Star City", State: "CA", Zip: "90001"},
	},
	Jobs: []Job{
		{Title: "Engineer", Company: "Acme Corp", Location: Address{Street: "1 Acme Way", City: "Metropolis", State: "NY", Zip: "10001"}, Years: 5},
		{Title: "Manager", Company: "Globex", Location: Address{Street: "2 Globex Ave", City: "Gotham", State: "NJ", Zip: "07001"}, Years: 3},
		{Title: "CTO", Company: "Initech", Location: Address{Street: "3 Initech Blvd", City: "Star City", State: "CA", Zip: "90001"}, Years: 2},
	},
	Metadata: map[string]string{
		"Department": "Engineering",
		"Level":      "Senior",
		"Status":     "Active",
		"Lang":       "Go",
		"Timezone":   "UTC-5",
	},
	Active: true,
	Scores: []int{98, 87, 92, 88, 76, 95, 89, 91, 85, 90, 93, 97, 99, 100, 82, 84, 86, 88, 90, 92},
	Tags:   []string{"golang", "backend", "cloud", "devops", "microservices", "api", "performance", "testing", "benchmark", "template"},
}

// Simple template
const simpleTemplate = `Name: {{.Name}}
Age: {{.Age}}
Email: {{.Email}}
Active: {{.Active}}
Tags: {{range .Tags}}{{.}}, {{end}}
Friends: {{range .Friends}}{{.}}, {{end}}
First Address: {{with index .Addresses 0}}{{.Street}}, {{.City}}, {{.State}} {{.Zip}}{{end}}
First Job: {{with index .Jobs 0}}{{.Title}} at {{.Company}} ({{.Years}} years){{end}}
Department: {{index .Metadata "Department"}}
Score Count: {{len .Scores}}
`

// Complex template with loops and conditionals
const complexTemplate = `
{{.Name}}'s Extended Profile
========================================
Age: {{.Age}}
Email: {{.Email}}
Status: {{if .Active}}Active{{else}}Inactive{{end}}
Department: {{index .Metadata "Department"}}
Level: {{index .Metadata "Level"}}
Timezone: {{index .Metadata "Timezone"}}

Addresses:
{{range $i, $addr := .Addresses}}  {{$i}}. {{$addr.Street}}, {{$addr.City}}, {{$addr.State}} {{$addr.Zip}}
{{end}}

Jobs:
{{range .Jobs}}  - {{.Title}} at {{.Company}} ({{.Years}} years), Location: {{.Location.City}}, {{.Location.State}}
{{end}}

Friends ({{len .Friends}}):
{{range .Friends}}  - {{.}}
{{end}}

Tags:
{{range .Tags}}[{{.}}] {{end}}

Scores:
{{range $i, $score := .Scores}}{{$score}}{{if lt $i (sub1 (len $.Scores))}}, {{end}}{{end}}

{{if gt .Age 18}}This person is an adult{{else}}This person is a minor{{end}}

{{if .Active}}
  {{if gt (len .Jobs) 2}}Highly experienced professional!{{else}}Still growing career.{{end}}
{{else}}
  Not currently active.
{{end}}

{{define "sub1"}}{{- $n := . -}}{{$n}}{{end}}
`

// Helper function for sub1 (Go templates don't have arithmetic by default)
func sub1(n int) int { return n - 1 }

// Benchmarks for text/template
func BenchmarkTextTemplateSimple(b *testing.B) {
	tmpl, err := text_template.New("simple").Funcs(map[string]interface{}{
		"sub1": sub1,
	}).Parse(simpleTemplate)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tmpl.Execute(io.Discard, testPerson)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTextTemplateComplex(b *testing.B) {
	tmpl, err := text_template.New("complex").Funcs(map[string]interface{}{
		"sub1": sub1,
	}).Parse(complexTemplate)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tmpl.Execute(io.Discard, testPerson)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmarks for html/template
func BenchmarkHTMLTemplateSimple(b *testing.B) {
	tmpl, err := html_template.New("simple").Funcs(map[string]interface{}{
		"sub1": sub1,
	}).Parse(simpleTemplate)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tmpl.Execute(io.Discard, testPerson)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTMLTemplateComplex(b *testing.B) {
	tmpl, err := html_template.New("complex").Funcs(map[string]interface{}{
		"sub1": sub1,
	}).Parse(complexTemplate)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tmpl.Execute(io.Discard, testPerson)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func main() {
	// This empty main function is required to make 'go test' work
	// To run benchmarks: go test -bench=. -benchmem
}
