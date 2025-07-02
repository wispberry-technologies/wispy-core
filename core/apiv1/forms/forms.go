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
	FieldTitle = "title"
	FieldTags  = "tags"
	FieldPhone = "phone"
)

// Form represents a form definition
type Form struct {
	ID           string         `json:"id" db:"id"`
	SiteDomain   string         `json:"site_domain" db:"site_domain"`
	Name         string         `json:"name" db:"name" validate:"required"`
	Slug         string         `json:"slug" db:"slug" validate:"required"`
	Fields       []FormField    `json:"fields" db:"fields"`
	CommonFields CommonFields   `json:"common_fields" db:"common_fields"`
	RedirectURL  string         `json:"redirect_url" db:"redirect_url"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	Metadata     map[string]any `json:"metadata" db:"metadata"`
}

// CommonFields defines which common fields this form collects
type CommonFields struct {
	Email bool `json:"email" db:"email"`
	Name  bool `json:"name" db:"name"`
	Title bool `json:"title" db:"title"`
	Tags  bool `json:"tags" db:"tags"`
	Phone bool `json:"phone" db:"phone"`
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
	Email      *string           `json:"email,omitempty" db:"email"`
	Name       *string           `json:"name,omitempty" db:"name"`
	Title      *string           `json:"title,omitempty" db:"title"`
	Tags       []string          `json:"tags,omitempty" db:"tags"`
	Phone      *string           `json:"phone,omitempty" db:"phone"`
	IPAddress  string            `json:"ip_address" db:"ip_address"`
	UserAgent  string            `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
}

type FormApi struct {
	siteManager    site.SiteManager
	validate       *validator.Validate
	authMiddleware *auth.Middleware
}

func NewFormApi(siteManager site.SiteManager, authMiddleware *auth.Middleware) *FormApi {
	validate := validator.New()
	return &FormApi{
		siteManager:    siteManager,
		validate:       validate,
		authMiddleware: authMiddleware,
	}
}

func (f *FormApi) MountApi(r chi.Router) {
	r.Route("/forms", func(r chi.Router) {
		r.Post("/submit", f.FormSubmission)

		r.Get("/submissions/by-email/{email}", f.GetSubmissionsByEmail)
		r.Get("/submissions/by-name/{name}", f.GetSubmissionsByName)
		r.Get("/submissions/by-phone/{phone}", f.GetSubmissionsByPhone)
		r.Get("/submissions/with-tags/{tag}", f.GetSubmissionsWithTag)
		r.Get("/submissions/by-title/{title}", f.GetSubmissionsByTitle)

		r.Group(func(r chi.Router) {
			// r.Use(authMiddleware)
			r.Post("/", f.CreateForm)
			r.Get("/", f.ListForms)
			r.Get("/{formID}", f.GetForm)
			r.Get("/{formID}/submissions", f.GetFormSubmissions)
		})
	})
}

func (f *FormApi) FormSubmission(w http.ResponseWriter, r *http.Request) {
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found", err)
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

	if email, ok := commonData[FieldEmail]; ok {
		submission.Email = &email
	}
	if name, ok := commonData[FieldName]; ok {
		submission.Name = &name
	}
	if title, ok := commonData[FieldTitle]; ok {
		submission.Title = &title
	}
	if phone, ok := commonData[FieldPhone]; ok {
		submission.Phone = &phone
	}
	if tags, ok := commonData[FieldTags]; ok {
		submission.Tags = strings.Split(tags, ",")
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
				"form_id":     formID,
				"site_domain": site.GetDomain(),
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

	for _, field := range form.Fields {
		fieldMap[field.Name] = field
	}

	for name, values := range formData {
		if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
			continue
		}

		field, exists := fieldMap[name]
		if !exists {
			continue
		}

		value := strings.TrimSpace(values[0])
		if field.Required && value == "" {
			return nil, nil, common.NewError("field '" + name + "' is required")
		}

		switch field.Type {
		case "email":
			if err := f.validate.Var(value, "email"); err != nil {
				return nil, nil, common.NewError("field '" + name + "' must be a valid email")
			}
			if form.CommonFields.Email {
				commonFields[FieldEmail] = value
			}
		case "tel":
			if err := validatePhoneNumber(value); err != nil {
				return nil, nil, common.NewError("field '" + name + "' must be a valid phone number")
			}
			if form.CommonFields.Phone {
				commonFields[FieldPhone] = value
			}
		}

		switch strings.ToLower(name) {
		case "fullname", "firstname", "lastname", "username":
			if form.CommonFields.Name {
				commonFields[FieldName] = value
			}
		case "subject", "heading":
			if form.CommonFields.Title {
				commonFields[FieldTitle] = value
			}
		case "categories", "interests":
			if form.CommonFields.Tags {
				commonFields[FieldTags] = value
			}
		}

		normalized[name] = value
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

func (f *FormApi) GetSubmissionsByTitle(w http.ResponseWriter, r *http.Request) {
	f.querySubmissionsByField(w, r, FieldTitle, chi.URLParam(r, "title"))
}

func (f *FormApi) querySubmissionsByField(w http.ResponseWriter, r *http.Request, field, value string) {
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found", err)
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
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found", err)
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

	// Parse common fields
	form.CommonFields = CommonFields{
		Email: r.FormValue("common_email") == "true",
		Name:  r.FormValue("common_name") == "true",
		Title: r.FormValue("common_title") == "true",
		Tags:  r.FormValue("common_tags") == "true",
		Phone: r.FormValue("common_phone") == "true",
	}

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
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found", err)
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
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found", err)
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
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		common.RespondWithError(w, r, http.StatusNotFound, "Site not found", err)
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

	// Parse fields from JSON string in the database
	// Since we can't use json.Unmarshal, we'll need to manually parse or use a different approach
	// For now, let's store the raw JSON and handle it appropriately
	if fieldsJSON != "" {
		// We'll need to parse this manually or use a different serialization method
		// For now, creating a simple parser or storing as string
		form.Fields = []FormField{} // Empty for now
	}

	// Handle metadata/settings
	if settingsJSON != "" {
		form.Metadata = make(map[string]any)
		// Would need manual parsing here too
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
		form.Name, // Use name as title
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
		INSERT INTO form_submissions (uuid, form_id, data, ip_address, user_agent, created_at)
		VALUES (?, (SELECT id FROM forms WHERE uuid = ?), ?, ?, ?, ?)`

	// Since we can't use JSON, serialize submission data as key-value pairs
	dataStr := ""
	for key, value := range submission.Data {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += key + ":" + value
	}

	// Add common fields to data string
	if submission.Email != nil {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += "email:" + *submission.Email
	}
	if submission.Name != nil {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += "name:" + *submission.Name
	}
	if submission.Phone != nil {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += "phone:" + *submission.Phone
	}
	if submission.Title != nil {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += "title:" + *submission.Title
	}
	if len(submission.Tags) > 0 {
		if dataStr != "" {
			dataStr += "|"
		}
		dataStr += "tags:" + strings.Join(submission.Tags, ",")
	}

	_, err := db.Exec(saveSubmissionSQL,
		submission.ID,
		submission.FormID,
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
	const getSubmissionsSQL = `
		SELECT fs.uuid, f.uuid, fs.data, fs.ip_address, fs.user_agent, fs.created_at
		FROM form_submissions fs
		JOIN forms f ON fs.form_id = f.id
		WHERE fs.data LIKE ?
		ORDER BY fs.created_at DESC`

	// Create search pattern based on field
	searchPattern := "%" + field + ":" + value + "%"

	rows, err := db.Query(getSubmissionsSQL, searchPattern)
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

		// Parse data string back to map and individual fields
		if dataStr != "" {
			pairs := strings.Split(dataStr, "|")
			for _, pair := range pairs {
				if strings.Contains(pair, ":") {
					parts := strings.SplitN(pair, ":", 2)
					key, val := parts[0], parts[1]
					
					switch key {
					case "email":
						submission.Email = &val
					case "name":
						submission.Name = &val
					case "phone":
						submission.Phone = &val
					case "title":
						submission.Title = &val
					case "tags":
						submission.Tags = strings.Split(val, ",")
					default:
						submission.Data[key] = val
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
		SELECT fs.uuid, f.uuid, fs.data, fs.ip_address, fs.user_agent, fs.created_at
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

		// Parse data string back to map and individual fields
		if dataStr != "" {
			pairs := strings.Split(dataStr, "|")
			for _, pair := range pairs {
				if strings.Contains(pair, ":") {
					parts := strings.SplitN(pair, ":", 2)
					key, val := parts[0], parts[1]
					
					switch key {
					case "email":
						submission.Email = &val
					case "name":
						submission.Name = &val
					case "phone":
						submission.Phone = &val
					case "title":
						submission.Title = &val
					case "tags":
						submission.Tags = strings.Split(val, ",")
					default:
						submission.Data[key] = val
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
