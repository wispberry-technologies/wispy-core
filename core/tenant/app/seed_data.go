package app

import (
	"context"
	"fmt"
	"time"
	"wispy-core/common"
)

// SeedSampleData seeds the database with sample data for testing
func (dp *CMSDataProvider) SeedSampleData(ctx context.Context) error {
	db, err := dp.GetFormsDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	defer db.Close()

	// Check if sample data already exists
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM forms").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		common.Info("Sample data already exists, skipping seeding")
		return nil
	}

	// Sample forms data
	forms := []struct {
		uuid        string
		name        string
		title       string
		description string
		fields      string
		settings    string
	}{
		{
			uuid:        "contact-form-001",
			name:        "contact_form",
			title:       "Contact Form",
			description: "General contact inquiries",
			fields:      `[{"type": "text", "name": "first_name", "label": "First Name", "required": true}, {"type": "text", "name": "last_name", "label": "Last Name", "required": true}, {"type": "email", "name": "email", "label": "Email", "required": true}, {"type": "text", "name": "subject", "label": "Subject", "required": true}, {"type": "textarea", "name": "message", "label": "Message", "required": true}]`,
			settings:    `{"confirmation_message": "Thank you for your message. We'll get back to you soon!"}`,
		},
		{
			uuid:        "newsletter-form-001",
			name:        "newsletter_signup",
			title:       "Newsletter Signup",
			description: "Email collection form",
			fields:      `[{"type": "email", "name": "email", "label": "Email Address", "required": true}]`,
			settings:    `{"confirmation_message": "Thank you for subscribing to our newsletter!"}`,
		},
		{
			uuid:        "feedback-form-001",
			name:        "feedback_form",
			title:       "Feedback Form",
			description: "Customer feedback collection",
			fields:      `[{"type": "text", "name": "name", "label": "Name", "required": false}, {"type": "email", "name": "email", "label": "Email", "required": false}, {"type": "select", "name": "rating", "label": "Rating", "options": ["5 - Excellent", "4 - Good", "3 - Average", "2 - Poor", "1 - Terrible"], "required": true}, {"type": "textarea", "name": "feedback", "label": "Feedback", "required": true}]`,
			settings:    `{"confirmation_message": "Thank you for your feedback!"}`,
		},
	}

	// Insert sample forms
	for _, form := range forms {
		_, err := db.ExecContext(ctx, `
			INSERT INTO forms (uuid, name, title, description, fields, settings, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, form.uuid, form.name, form.title, form.description, form.fields, form.settings, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert form %s: %w", form.name, err)
		}
	}

	// Get form IDs for submissions
	formIDs := make(map[string]int)
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM forms")
	if err != nil {
		return fmt.Errorf("failed to query forms: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		formIDs[name] = id
	}

	// Sample submissions data
	submissions := []struct {
		uuid      string
		formName  string
		firstName string
		lastName  string
		email     string
		subject   string
		message   string
		data      string
		createdAt time.Time
	}{
		{
			uuid:      "sub-001",
			formName:  "contact_form",
			firstName: "John",
			lastName:  "Doe",
			email:     "john@example.com",
			subject:   "Website inquiry about services",
			message:   "I'm interested in your web development services. Can you please provide more information about your pricing and timeline?",
			data:      `{"first_name": "John", "last_name": "Doe", "email": "john@example.com", "subject": "Website inquiry about services", "message": "I'm interested in your web development services. Can you please provide more information about your pricing and timeline?"}`,
			createdAt: time.Now().Add(-2 * time.Hour),
		},
		{
			uuid:      "sub-002",
			formName:  "newsletter_signup",
			firstName: "Sarah",
			lastName:  "Wilson",
			email:     "sarah@example.com",
			subject:   "",
			message:   "Newsletter subscription",
			data:      `{"email": "sarah@example.com"}`,
			createdAt: time.Now().Add(-4 * time.Hour),
		},
		{
			uuid:      "sub-003",
			formName:  "feedback_form",
			firstName: "Mike",
			lastName:  "Johnson",
			email:     "mike@example.com",
			subject:   "",
			message:   "Great service, very satisfied!",
			data:      `{"name": "Mike Johnson", "email": "mike@example.com", "rating": "5 - Excellent", "feedback": "Great service, very satisfied!"}`,
			createdAt: time.Now().Add(-1 * 24 * time.Hour),
		},
		{
			uuid:      "sub-004",
			formName:  "contact_form",
			firstName: "Emma",
			lastName:  "Davis",
			email:     "emma@example.com",
			subject:   "Support request",
			message:   "I need help with my account settings.",
			data:      `{"first_name": "Emma", "last_name": "Davis", "email": "emma@example.com", "subject": "Support request", "message": "I need help with my account settings."}`,
			createdAt: time.Now().Add(-6 * time.Hour),
		},
		{
			uuid:      "sub-005",
			formName:  "newsletter_signup",
			firstName: "Tom",
			lastName:  "Brown",
			email:     "tom@example.com",
			subject:   "",
			message:   "Newsletter subscription",
			data:      `{"email": "tom@example.com"}`,
			createdAt: time.Now().Add(-12 * time.Hour),
		},
	}

	// Insert sample submissions
	for _, submission := range submissions {
		formID, exists := formIDs[submission.formName]
		if !exists {
			continue
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO form_submissions (uuid, form_id, first_name, last_name, email, subject, message, data, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, submission.uuid, formID, submission.firstName, submission.lastName, submission.email, submission.subject, submission.message, submission.data, submission.createdAt)
		if err != nil {
			return fmt.Errorf("failed to insert submission %s: %w", submission.uuid, err)
		}
	}

	common.Info("Sample data seeded successfully")
	return nil
}

// SeedSampleDataForDomain is a helper function to seed sample data for a specific domain
func SeedSampleDataForDomain(domain string) error {
	dataProvider := NewCMSDataProvider(domain)
	return dataProvider.SeedSampleData(context.Background())
}
