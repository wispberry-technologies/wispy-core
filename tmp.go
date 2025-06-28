package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Sample struct to decode into
type User struct {
	Name    string  `json:"name" form:"name"`
	Email   string  `json:"email" form:"email"`
	Age     int     `json:"age" form:"age"`
	Active  bool    `json:"active" form:"active"`
	Balance float64 `json:"balance" form:"balance"`
}

func main() {
	// Sample data in both formats
	jsonData := `{"name":"John Doe","email":"john@example.com","age":30,"active":true,"balance":1250.75}`
	formData := "name=John+Doe&email=john%40example.com&age=30&active=true&balance=1250.75"

	// Run benchmarks
	jsonTime, jsonErr := benchmarkJSONDecoding(jsonData)
	formTime, formErr := benchmarkFormDecoding(formData)

	// Print results
	fmt.Println("=== Decoding Performance Comparison ===")
	fmt.Printf("JSON decoding time: %v\n", jsonTime)
	if jsonErr != nil {
		fmt.Printf("JSON decoding error: %v\n", jsonErr)
	}

	fmt.Printf("Form decoding time: %v\n", formTime)
	if formErr != nil {
		fmt.Printf("Form decoding error: %v\n", formErr)
	}

	fmt.Printf("\nForm decoding is %.2fx %s than JSON decoding\n",
		calculateRatio(jsonTime, formTime),
		compareSpeed(jsonTime, formTime))
}

func benchmarkJSONDecoding(data string) (time.Duration, error) {
	var user User
	start := time.Now()

	decoder := json.NewDecoder(strings.NewReader(data))
	err := decoder.Decode(&user)

	return time.Since(start), err
}

func benchmarkFormDecoding(data string) (time.Duration, error) {
	var user User
	start := time.Now()

	values, err := url.ParseQuery(data)
	if err != nil {
		return time.Since(start), err
	}

	// Manually map form values to struct
	// In a real application, you might use a library like gorilla/schema
	user.Name = values.Get("name")
	user.Email = values.Get("email")
	user.Age, err = parseInt(values.Get("age"))
	if err != nil {
		return time.Since(start), err
	}
	user.Active, err = parseBool(values.Get("active"))
	if err != nil {
		return time.Since(start), err
	}
	user.Balance, err = parseFloat(values.Get("balance"))
	if err != nil {
		return time.Since(start), err
	}

	return time.Since(start), nil
}

// Helper functions for form value parsing
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "t":
		return true, nil
	case "false", "0", "f":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func calculateRatio(jsonTime, formTime time.Duration) float64 {
	if jsonTime > formTime {
		return float64(jsonTime) / float64(formTime)
	}
	return float64(formTime) / float64(jsonTime)
}

func compareSpeed(jsonTime, formTime time.Duration) string {
	if jsonTime > formTime {
		return "faster"
	}
	return "slower"
}
