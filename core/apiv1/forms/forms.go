package forms

import (
	"database/sql"
	"encoding/json"
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
	ID          string            `json:"id" db:"id"`
	FormID      string            `json:"form_id" db:"form_id"`
	SiteDomain  string            `json:"site_domain" db:"site_domain"`
	Data        map[string]string `json:"data" db:"data"`
	Email     *string           `json:"email,omitempty" db:"email"`
	Name      *string           `json:"name,omitempty" db:"name"`
	Title     *string           `json:"title,omitempty" db:"title"`
	Tags      []string          `json:"tags,omitempty" db:"tags"`
	Phone     *string           `json:"phone,omitempty" db:"phone"`
	IPAddress string            `json:"ip_address" db:"ip_address"`
	UserAgent string            `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
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
		ID:          uuid.New().String(),
		FormID:      formID,
		SiteDomain:  site.GetDomain(),
		Data:        submissionData,
		IPAddress:   common.GetIPAddress(r),
		UserAgent:   r.UserAgent(),
		CreatedAt:   time.Now(),
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

	var form Form
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		common.RespondWithError(w, r, http.StatusBadRequest, "Invalid form data", err)
		return
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
	// Database implementation would go here
	return Form{}, nil
}

func (f *FormApi) saveForm(db *sql.DB, form Form) error {
	// Database implementation would go here
	return nil
}

func (f *FormApi) getAllForms(db *sql.DB, siteID string) ([]Form, error) {
	// Database implementation would go here
	return []Form{}, nil
}

func (f *FormApi) saveSubmission(db *sql.DB, submission FormSubmission) error {
	// Database implementation would go here
	return nil
}

func (f *FormApi) getSubmissionsByField(db *sql.DB, siteID, field, value string) ([]FormSubmission, error) {
	// Database implementation would go here
	return []FormSubmission{}, nil
}

func (f *FormApi) getFormSubmissions(db *sql.DB, siteID, formID string) ([]FormSubmission, error) {
	// Database implementation would go here
	return []FormSubmission{}, nil
}

func validatePhoneNumber(phone string) error {
	if len(phone) < 5 || len(phone) > 20 {
		return common.NewError("phone: invalid length")
	}
	return nil
}
