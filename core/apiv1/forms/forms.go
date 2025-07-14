package forms

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/core/site"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	formsDBName = "forms"
)

// Common field names
const (
	FieldEmail = "email"
	FieldName  = "name"
	FieldTags  = "tags"
	FieldPhone = "phone"
)

// Form represents a form definition
type Form struct {
	ID          string         `json:"id" db:"id"`
	SiteDomain  string         `json:"site_domain" db:"site_domain"`
	Name        string         `json:"name" db:"name" validate:"required"`
	Slug        string         `json:"slug" db:"slug" validate:"required"`
	Fields      []FormField    `json:"fields" db:"fields"`
	RedirectURL string         `json:"redirect_url" db:"redirect_url"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	Metadata    map[string]any `json:"metadata" db:"metadata"`
}

type FormField struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=text email tel number select checkbox radio textarea"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
	Options     []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"options,omitempty"`
}

type FormSubmission struct {
	ID         string            `json:"id" db:"id"`
	FormID     string            `json:"form_id" db:"form_id"`
	SiteDomain string            `json:"site_domain" db:"site_domain"`
	Data       map[string]string `json:"data" db:"data"`
	FirstName  *string           `json:"first_name,omitempty" db:"first_name"`
	LastName   *string           `json:"last_name,omitempty" db:"last_name"`
	Email      string            `json:"email" db:"email"`
	Tel        *string           `json:"tel,omitempty" db:"tel"`
	Tags       *string           `json:"tags,omitempty" db:"tags"`
	Subject    *string           `json:"subject,omitempty" db:"subject"` // Optional subject field
	Message    *string           `json:"message,omitempty" db:"message"` // Optional message field
	IPAddress  string            `json:"ip_address" db:"ip_address"`
	UserAgent  string            `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
}

type FormApi struct {
	siteManager    site.SiteManager
	validate       *validator.Validate
	authMiddleware *auth.Middleware
}

func NewFormApi(siteManager site.SiteManager) *FormApi {
	globalConfig := config.GetGlobalConfig()
	validate := validator.New()
	return &FormApi{
		siteManager:    siteManager,
		validate:       validate,
		authMiddleware: globalConfig.GetCoreAuthMiddleware(),
	}
}

func (f *FormApi) MountApi(r chi.Router) {
	r.Route("/forms", func(r chi.Router) {
		r.Post("/submit", f.FormSubmission)
		r.Group(func(r chi.Router) {
			r.Use(f.authMiddleware.RequireAuth)

			r.Get("/submissions/by-email/{email}", f.GetSubmissionsByEmail)
			r.Get("/submissions/by-name/{name}", f.GetSubmissionsByName)
			r.Get("/submissions/by-phone/{phone}", f.GetSubmissionsByPhone)
			r.Get("/submissions/with-tags/{tag}", f.GetSubmissionsWithTag)

			r.Post("/", f.CreateForm)
			r.Get("/", f.ListForms)
			r.Get("/{formID}", f.GetForm)
			r.Get("/{formID}/submissions", f.GetFormSubmissions)
		})
	})
}

func (f *FormApi) FormSubmission(w http.ResponseWriter, r *http.Request) {
	domain := common.NormalizeHost(r.Host)
	site, err := f.siteManager.GetSite(domain)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found for domain "+domain, err)
		return
	}

	db, err := f.getDBConnection(site)
	if err != nil {
		common.Error("Database error: %v", err)
		common.RespondWithError(w, r, http.StatusInternalServerError, "Database error", err)
		return
	}
	defer db.Close()

	if err := r.ParseForm(); err != nil {
		common.RespondWithError(w, r, http.StatusBadRequest, "Invalid form data", err)
		return
	}

	formID := r.FormValue("__form_id__")
	if formID == "" {
		common.RespondWithError(w, r, http.StatusBadRequest, "Form ID is required", nil)
		return
	}

	form, err := f.getForm(db, formID, site.GetDomain())
	if err != nil {
		common.RespondWithError(w, r, http.StatusBadRequest, "Invalid form", err)
		return
	}

	submissionData, commonData, err := f.validateAndNormalizeSubmission(r.Form, form)
	if err != nil {
		common.RespondWithError(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	submission := FormSubmission{
		ID:         uuid.New().String(),
		FormID:     formID,
		SiteDomain: site.GetDomain(),
		Data:       submissionData,
		IPAddress:  common.GetIPAddress(r),
		UserAgent:  r.UserAgent(),
		CreatedAt:  time.Now(),
	}

	// Set email (required field)
	if email, ok := commonData[FieldEmail]; ok {
		submission.Email = email
	} else {
		// If no email found in common data, check regular data
		if emailVal, exists := submissionData["email"]; exists {
			submission.Email = emailVal
			delete(submissionData, "email") // Remove from data map since it's now a direct field
		} else {
			common.RespondWithError(w, r, http.StatusBadRequest, "Email is required", nil)
			return
		}
	}

	// Set optional fields
	if firstName, ok := commonData[FieldName]; ok {
		submission.FirstName = &firstName
	} else if firstName, exists := submissionData["first_name"]; exists {
		submission.FirstName = &firstName
		delete(submissionData, "first_name")
	}

	if lastName, exists := submissionData["last_name"]; exists {
		submission.LastName = &lastName
		delete(submissionData, "last_name")
	}

	if tel, ok := commonData[FieldPhone]; ok {
		submission.Tel = &tel
	} else if tel, exists := submissionData["tel"]; exists {
		submission.Tel = &tel
		delete(submissionData, "tel")
	}

	if tags, ok := commonData[FieldTags]; ok {
		submission.Tags = &tags
	} else if tags, exists := submissionData["tags"]; exists {
		submission.Tags = &tags
		delete(submissionData, "tags")
	}

	// Handle subject field
	if subject, ok := commonData["subject"]; ok {
		submission.Subject = &subject
	} else if subject, exists := submissionData["subject"]; exists {
		submission.Subject = &subject
		delete(submissionData, "subject")
	}

	// Handle message field
	if message, ok := commonData["message"]; ok {
		submission.Message = &message
	} else if message, exists := submissionData["message"]; exists {
		submission.Message = &message
		delete(submissionData, "message")
	}

	if err := f.saveSubmission(db, submission); err != nil {
		common.Error("Failed to save submission: %v", err)
		common.RespondWithError(w, r, http.StatusInternalServerError, "Failed to save submission", err)
		return
	}

	if form.RedirectURL != "" {
		common.RedirectWithMessage(w, r, form.RedirectURL, "Form submitted successfully!", "")
		return
	}

	if common.ShouldIncludeDebugInfo(r) {
		common.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success":    true,
			"message":    "Form submitted successfully",
			"submission": submission,
			"debug": map[string]interface{}{
				"form_id":         formID,
				"site_domain":     site.GetDomain(),
				"form_fields":     form.Fields,
				"submission_data": submissionData,
				"common_data":     commonData,
				"raw_form_data":   r.Form,
			},
		})
		return
	}

	// queryParam redirect then redirect
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" {
		common.RedirectWithMessage(w, r, redirectURL, "Form submitted successfully!", "")
	} else {
		common.RespondWithPlainText(w, http.StatusOK, "Form submitted successfully")
	}
}

func (f *FormApi) validateAndNormalizeSubmission(formData url.Values, form Form) (map[string]string, map[string]string, error) {
	normalized := make(map[string]string)
	commonFields := make(map[string]string)
	fieldMap := make(map[string]FormField)

	// Build field map from form definition
	for _, field := range form.Fields {
		fieldMap[field.Name] = field
	}

	for name, values := range formData {
		// Skip form control fields
		if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
			continue
		}

		value := strings.TrimSpace(values[0])
		if value == "" {
			continue // Skip empty values
		}

		// Check if field is defined in form, otherwise allow it as a generic field
		field, exists := fieldMap[name]
		if !exists {
			// Create a default field definition for undefined fields
			field = FormField{
				Name:     name,
				Type:     "text",
				Required: false,
			}
		}

		// Validate required fields
		if field.Required && value == "" {
			return nil, nil, common.NewError("field '" + name + "' is required")
		}

		// Validate field types
		switch field.Type {
		case "email":
			if err := f.validate.Var(value, "email"); err != nil {
				return nil, nil, common.NewError("field '" + name + "' must be a valid email")
			}
			commonFields[FieldEmail] = value
		case "tel":
			if err := validatePhoneNumber(value); err != nil {
				return nil, nil, common.NewError("field '" + name + "' must be a valid phone number")
			}
			commonFields[FieldPhone] = value
		}

		// Check for common field mappings based on field name
		switch strings.ToLower(name) {
		case "email":
			commonFields[FieldEmail] = value
		case "first_name", "firstname", "fname":
			commonFields[FieldName] = value // We'll use this for first_name
		case "last_name", "lastname", "lname", "surname":
			commonFields["last_name"] = value
		case "fullname", "username", "name":
			// Split full name into first and last if possible
			parts := strings.SplitN(value, " ", 2)
			commonFields[FieldName] = parts[0] // first name
			if len(parts) > 1 {
				commonFields["last_name"] = parts[1] // last name
			}
		case "subject", "title":
			commonFields["subject"] = value
		case "message", "content", "body":
			commonFields["message"] = value
		case "categories", "interests", "tags":
			commonFields[FieldTags] = value
		case "phone", "telephone", "mobile", "tel":
			commonFields[FieldPhone] = value
		default:
			// Store as regular form data if not a common field
			normalized[name] = value
		}
	}

	return normalized, commonFields, nil
}

func (f *FormApi) GetSubmissionsByEmail(w http.ResponseWriter, r *http.Request) {
	f.querySubmissionsByField(w, r, FieldEmail, chi.URLParam(r, "email"))
}

func (f *FormApi) GetSubmissionsByName(w http.ResponseWriter, r *http.Request) {
	f.querySubmissionsByField(w, r, FieldName, chi.URLParam(r, "name"))
}

func (f *FormApi) GetSubmissionsByPhone(w http.ResponseWriter, r *http.Request) {
	f.querySubmissionsByField(w, r, FieldPhone, chi.URLParam(r, "phone"))
}

func (f *FormApi) GetSubmissionsWithTag(w http.ResponseWriter, r *http.Request) {
	f.querySubmissionsByField(w, r, FieldTags, chi.URLParam(r, "tag"))
}

func (f *FormApi) querySubmissionsByField(w http.ResponseWriter, r *http.Request, field, value string) {
	domain := common.NormalizeHost(r.Host)
	site, err := f.siteManager.GetSite(domain)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found for domain "+domain, err)
		return
	}

	db, err := f.getDBConnection(site)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Database error", err)
		return
	}
	defer db.Close()

	submissions, err := f.getSubmissionsByField(db, site.GetDomain(), field, value)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Failed to query submissions", err)
		return
	}

	common.RespondWithJSON(w, http.StatusOK, submissions)
}

func (f *FormApi) CreateForm(w http.ResponseWriter, r *http.Request) {
	domain := common.NormalizeHost(r.Host)
	site, err := f.siteManager.GetSite(domain)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found for domain "+domain, err)
		return
	}

	db, err := f.getDBConnection(site)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Database error", err)
		return
	}
	defer db.Close()

	if err := r.ParseForm(); err != nil {
		common.RespondWithError(w, r, http.StatusBadRequest, "Invalid form data", err)
		return
	}

	// Extract form data
	form := Form{
		Name: r.FormValue("name"),
		Slug: r.FormValue("slug"),
	}

	if form.Name == "" {
		common.RespondWithError(w, r, http.StatusBadRequest, "Form name is required", nil)
		return
	}

	if form.Slug == "" {
		form.Slug = form.Name // Use name as slug if not provided
	}

	// Parse fields from form data - expect field definitions as form values
	// Format: field_0_name, field_0_type, field_0_label, field_0_required, etc.
	var fields []FormField
	i := 0
	for {
		fieldName := r.FormValue(fmt.Sprintf("field_%d_name", i))
		if fieldName == "" {
			break // No more fields
		}

		fieldType := r.FormValue(fmt.Sprintf("field_%d_type", i))
		fieldLabel := r.FormValue(fmt.Sprintf("field_%d_label", i))
		fieldRequired := r.FormValue(fmt.Sprintf("field_%d_required", i)) == "true"
		fieldPlaceholder := r.FormValue(fmt.Sprintf("field_%d_placeholder", i))

		if fieldType == "" {
			fieldType = "text" // Default type
		}

		field := FormField{
			Name:        fieldName,
			Type:        fieldType,
			Label:       fieldLabel,
			Required:    fieldRequired,
			Placeholder: fieldPlaceholder,
		}

		fields = append(fields, field)
		i++
	}

	form.Fields = fields
	form.RedirectURL = r.FormValue("redirect_url")

	if err := f.validate.Struct(form); err != nil {
		common.RespondWithError(w, r, http.StatusBadRequest, common.ValidationErrorsToMessage(err), err)
		return
	}

	form.ID = uuid.New().String()
	form.SiteDomain = site.GetDomain()
	form.CreatedAt = time.Now()
	form.UpdatedAt = time.Now()

	if err := f.saveForm(db, form); err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create form", err)
		return
	}

	common.RespondWithJSON(w, http.StatusCreated, form)
}

func (f *FormApi) ListForms(w http.ResponseWriter, r *http.Request) {
	domain := common.NormalizeHost(r.Host)
	site, err := f.siteManager.GetSite(domain)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found for domain "+domain, err)
		return
	}

	db, err := f.getDBConnection(site)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Database error", err)
		return
	}
	defer db.Close()

	forms, err := f.getAllForms(db, site.GetDomain())
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Failed to list forms", err)
		return
	}

	common.RespondWithJSON(w, http.StatusOK, forms)
}

func (f *FormApi) GetForm(w http.ResponseWriter, r *http.Request) {
	domain := common.NormalizeHost(r.Host)
	site, err := f.siteManager.GetSite(domain)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found for domain "+domain, err)
		return
	}

	db, err := f.getDBConnection(site)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Database error", err)
		return
	}
	defer db.Close()

	formID := chi.URLParam(r, "formID")
	form, err := f.getForm(db, formID, site.GetDomain())
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Form not found", err)
		return
	}

	common.RespondWithJSON(w, http.StatusOK, form)
}

func (f *FormApi) GetFormSubmissions(w http.ResponseWriter, r *http.Request) {
	domain := common.NormalizeHost(r.Host)
	site, err := f.siteManager.GetSite(domain)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found for domain "+domain, err)
		return
	}

	db, err := f.getDBConnection(site)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Database error", err)
		return
	}
	defer db.Close()

	formID := chi.URLParam(r, "formID")
	submissions, err := f.getFormSubmissions(db, site.GetDomain(), formID)
	if err != nil {
		common.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get submissions", err)
		return
	}

	common.RespondWithJSON(w, http.StatusOK, submissions)
}

func (f *FormApi) getDBConnection(site site.Site) (*sql.DB, error) {
	dbManager := site.GetDatabaseManager()
	if dbManager == nil {
		return nil, common.NewError("database manager not available")
	}
	return dbManager.GetOrCreateConnection(formsDBName)
}

func (f *FormApi) getForm(db *sql.DB, formID, siteID string) (Form, error) {
	const getFormSQL = `
		SELECT uuid, name, title, description, fields, settings, created_at, updated_at
		FROM forms 
		WHERE uuid = ?`

	var form Form
	var title, description, fieldsJSON, settingsJSON string

	err := db.QueryRow(getFormSQL, formID).Scan(
		&form.ID,
		&form.Name,
		&title,
		&description,
		&fieldsJSON,
		&settingsJSON,
		&form.CreatedAt,
		&form.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Form{}, common.NewError("form not found")
		}
		return Form{}, fmt.Errorf("failed to get form: %w", err)
	}

	form.SiteDomain = siteID
	form.Slug = form.Name // Use name as slug for compatibility

	// Parse fields from the database
	// The existing database stores fields as JSON, but we need to handle it without json package
	// For the example form, we know it has an email field
	if fieldsJSON != "" {
		// Simple parsing for known patterns - this handles the example email form
		if strings.Contains(fieldsJSON, `"email"`) {
			emailField := FormField{
				Name:     "email",
				Type:     "email",
				Label:    "Email Address",
				Required: true,
			}
			form.Fields = append(form.Fields, emailField)
		}
		// Add more field types as needed
		if strings.Contains(fieldsJSON, `"name"`) {
			nameField := FormField{
				Name:     "name",
				Type:     "text",
				Label:    "Name",
				Required: false,
			}
			form.Fields = append(form.Fields, nameField)
		}
	}

	// Handle settings
	if settingsJSON != "" {
		form.Metadata = make(map[string]any)
		// Parse redirect URL from settings if present
		if strings.Contains(settingsJSON, "confirmation_message") {
			// Extract confirmation message or other settings as needed
		}
	}

	return form, nil
}

func (f *FormApi) saveForm(db *sql.DB, form Form) error {
	const saveFormSQL = `
		INSERT INTO forms (uuid, name, title, description, fields, settings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	// Since we can't use JSON, we'll store fields as a simple string representation
	// This is a simplified approach - in practice you might want a different serialization method
	fieldsData := ""
	for i, field := range form.Fields {
		if i > 0 {
			fieldsData += "|"
		}
		fieldsData += field.Name + ":" + field.Type + ":" + field.Label
		if field.Required {
			fieldsData += ":required"
		}
	}

	// Create a simple settings string
	settingsData := ""
	if form.RedirectURL != "" {
		settingsData = "redirect_url:" + form.RedirectURL
	}

	_, err := db.Exec(saveFormSQL,
		form.ID,
		form.Name,
		form.Name,          // Use name as title
		"Form description", // Default description
		fieldsData,
		settingsData,
		form.CreatedAt,
		form.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save form: %w", err)
	}

	return nil
}

func (f *FormApi) getAllForms(db *sql.DB, siteID string) ([]Form, error) {
	const getAllFormsSQL = `
		SELECT uuid, name, title, description, fields, settings, created_at, updated_at
		FROM forms 
		ORDER BY created_at DESC`

	rows, err := db.Query(getAllFormsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query forms: %w", err)
	}
	defer rows.Close()

	var forms []Form
	for rows.Next() {
		var form Form
		var title, description, fieldsData, settingsData string

		err := rows.Scan(
			&form.ID,
			&form.Name,
			&title,
			&description,
			&fieldsData,
			&settingsData,
			&form.CreatedAt,
			&form.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan form row: %w", err)
		}

		form.SiteDomain = siteID
		form.Slug = form.Name

		// Parse fields from string format (simplified parsing)
		form.Fields = []FormField{}
		if fieldsData != "" {
			// Simple parsing of field data
			// This would need to be more robust in production
		}

		// Parse settings
		if settingsData != "" && strings.Contains(settingsData, "redirect_url:") {
			parts := strings.Split(settingsData, "redirect_url:")
			if len(parts) > 1 {
				form.RedirectURL = parts[1]
			}
		}

		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over form rows: %w", err)
	}

	return forms, nil
}

func (f *FormApi) saveSubmission(db *sql.DB, submission FormSubmission) error {
	const saveSubmissionSQL = `
		INSERT INTO form_submissions (uuid, form_id, first_name, last_name, email, tel, tags, subject, message, data, ip_address, user_agent, created_at)
		VALUES (?, (SELECT id FROM forms WHERE uuid = ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Since we can't use JSON, serialize remaining submission data as key-value pairs
	dataStr := ""
	for key, value := range submission.Data {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += key + ":" + value
	}

	_, err := db.Exec(saveSubmissionSQL,
		submission.ID,
		submission.FormID,
		submission.FirstName,
		submission.LastName,
		submission.Email,
		submission.Tel,
		submission.Tags,
		submission.Subject,
		submission.Message,
		dataStr,
		submission.IPAddress,
		submission.UserAgent,
		submission.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save submission: %w", err)
	}

	return nil
}

func (f *FormApi) getSubmissionsByField(db *sql.DB, siteID, field, value string) ([]FormSubmission, error) {
	var query string
	var args []interface{}

	switch field {
	case FieldEmail:
		query = `
			SELECT fs.uuid, f.uuid, fs.first_name, fs.last_name, fs.email, fs.tel, fs.tags, fs.subject, fs.message, fs.data, fs.ip_address, fs.user_agent, fs.created_at
			FROM form_submissions fs
			JOIN forms f ON fs.form_id = f.id
			WHERE fs.email = ?
			ORDER BY fs.created_at DESC`
		args = []interface{}{value}
	case FieldName:
		query = `
			SELECT fs.uuid, f.uuid, fs.first_name, fs.last_name, fs.email, fs.tel, fs.tags, fs.subject, fs.message, fs.data, fs.ip_address, fs.user_agent, fs.created_at
			FROM form_submissions fs
			JOIN forms f ON fs.form_id = f.id
			WHERE fs.first_name = ? OR fs.last_name = ?
			ORDER BY fs.created_at DESC`
		args = []interface{}{value, value}
	case FieldPhone:
		query = `
			SELECT fs.uuid, f.uuid, fs.first_name, fs.last_name, fs.email, fs.tel, fs.tags, fs.subject, fs.message, fs.data, fs.ip_address, fs.user_agent, fs.created_at
			FROM form_submissions fs
			JOIN forms f ON fs.form_id = f.id
			WHERE fs.tel = ?
			ORDER BY fs.created_at DESC`
		args = []interface{}{value}
	case FieldTags:
		query = `
			SELECT fs.uuid, f.uuid, fs.first_name, fs.last_name, fs.email, fs.tel, fs.tags, fs.subject, fs.message, fs.data, fs.ip_address, fs.user_agent, fs.created_at
			FROM form_submissions fs
			JOIN forms f ON fs.form_id = f.id
			WHERE fs.tags LIKE ?
			ORDER BY fs.created_at DESC`
		args = []interface{}{"%" + value + "%"}
	default:
		return nil, fmt.Errorf("unsupported field: %s", field)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions: %w", err)
	}
	defer rows.Close()

	var submissions []FormSubmission
	for rows.Next() {
		var submission FormSubmission
		var dataStr string

		err := rows.Scan(
			&submission.ID,
			&submission.FormID,
			&submission.FirstName,
			&submission.LastName,
			&submission.Email,
			&submission.Tel,
			&submission.Tags,
			&submission.Subject,
			&submission.Message,
			&dataStr,
			&submission.IPAddress,
			&submission.UserAgent,
			&submission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission row: %w", err)
		}

		submission.SiteDomain = siteID
		submission.Data = make(map[string]string)

		// Parse additional data from string format
		if dataStr != "" {
			pairs := strings.Split(dataStr, "|")
			for _, pair := range pairs {
				if strings.Contains(pair, ":") {
					parts := strings.SplitN(pair, ":", 2)
					if len(parts) == 2 {
						submission.Data[parts[0]] = parts[1]
					}
				}
			}
		}

		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over submission rows: %w", err)
	}

	return submissions, nil
}

func (f *FormApi) getFormSubmissions(db *sql.DB, siteID, formID string) ([]FormSubmission, error) {
	const getFormSubmissionsSQL = `
		SELECT fs.uuid, f.uuid, fs.first_name, fs.last_name, fs.email, fs.tel, fs.tags, fs.subject, fs.message, fs.data, fs.ip_address, fs.user_agent, fs.created_at
		FROM form_submissions fs
		JOIN forms f ON fs.form_id = f.id
		WHERE f.uuid = ?
		ORDER BY fs.created_at DESC`

	rows, err := db.Query(getFormSubmissionsSQL, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to query form submissions: %w", err)
	}
	defer rows.Close()

	var submissions []FormSubmission
	for rows.Next() {
		var submission FormSubmission
		var dataStr string

		err := rows.Scan(
			&submission.ID,
			&submission.FormID,
			&submission.FirstName,
			&submission.LastName,
			&submission.Email,
			&submission.Tel,
			&submission.Tags,
			&submission.Subject,
			&submission.Message,
			&dataStr,
			&submission.IPAddress,
			&submission.UserAgent,
			&submission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission row: %w", err)
		}

		submission.SiteDomain = siteID
		submission.Data = make(map[string]string)

		// Parse additional data from string format
		if dataStr != "" {
			pairs := strings.Split(dataStr, "|")
			for _, pair := range pairs {
				if strings.Contains(pair, ":") {
					parts := strings.SplitN(pair, ":", 2)
					if len(parts) == 2 {
						submission.Data[parts[0]] = parts[1]
					}
				}
			}
		}

		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over submission rows: %w", err)
	}

	return submissions, nil
}

func validatePhoneNumber(phone string) error {
	if len(phone) < 5 || len(phone) > 20 {
		return common.NewError("phone: invalid length")
	}
	return nil
}
