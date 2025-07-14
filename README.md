# Wispy Core CMS - Project Guide

## Overview
Wispy Core is a modern, high-performance, multi-site content management system built with Go.

## CMS Design & Architecture

### Design Philosophy
Wispy CMS follows a **modern, atomic design system** approach using daisyUI v5 components. The interface is designed to be:
- **Content focused & minimal** - Focus on content and functionality without clutter
- **Responsive** - Works seamlessly across desktop, tablet, and mobile devices
- **Consistent** - Unified design language across all pages and components
- **Design & Style** - Inspired by Notion, Obsidian, and content focused applications
- **Accessible** - WCAG 2.1 AA compliant with proper semantic HTML
- **Fast** - Optimized for performance with minimal JavaScript

### Atomic Design System
The CMS uses a hierarchical component structure:

#### **Atoms** (`_data/design/systems/atoms/`)
- `button.html` - Primary, secondary, outline, and utility buttons
- `input.html` - Text, email, password, number inputs with validation states
- `label.html` - Form labels with required field indicators
- `alert.html` - Success, error, warning, and info notifications
- `badge.html` - Status indicators and tags
- `card.html` - Content containers with optional actions

#### **Components** (`_data/design/systems/components/`)
- `cms-navbar.html` - Main navigation with active page indicators
- `form-field.html` - Complete form field with label, input, and error handling
- `login-form.html` - Authentication form with error states
- `stats-dashboard.html` - Analytics widgets for dashboard
- `breadcrumb.html` - Navigation breadcrumbs
- `modal.html` - Dialog boxes and confirmations
- `loading.html` - Loading states and spinners
- `empty-state.html` - Placeholder content for empty data sets
- `table.html` - Data tables with sorting and actions

### Page Structure & Functionality

#### **Authentication Pages**

##### **Login Page** (`/wispy-cms/login`)
**Purpose**: Secure authentication entry point
**Design**: Clean, centered login form with branding
**Features**:
- Email/password authentication
- Remember me option
- Form validation with error states


#### **Main CMS Pages**

##### **Dashboard** (`/wispy-cms/dashboard`)
**Purpose**: Central hub and overview of CMS activity
**Design**: Card-based layout with key metrics and quick actions
**Features**:
- **TBD Section**: nothing yet

##### **Forms Management** (`/wispy-cms/forms`)
**Purpose**: Comprehensive form management interface
**Design**: List view with search, filter, and bulk actions
**Features**:
- **Search & Filter**: By name, status, date range, submission count
- **Forms List**: All forms with status, submission count, creation date
- **Status Indicators**: Active, draft, disabled states
- **Create Form**: Button to add new forms

##### **Form Submissions** (`/wispy-cms/forms/submissions`)
**Purpose**: View and manage all form submissions
**Design**: Table-based interface with filtering and bulk operations
**Features**:
- **Read/Unread States**: Visual indicators for submission status
- **Advanced Filtering**: By form, date range, read/unread status
- **Submissions Table**: Sortable columns (date, form, status, sender)
- **Individual Actions**: View details, mark as read, delete
- **Export Functions**: CSV export with custom field selection
- **Pagination**: Efficient loading of large datasets

##### **Settings** (`/wispy-cms/settings`)
**Purpose**: System configuration and preferences
**Design**: Tabbed interface with grouped settings
**Features**:
- **General Settings**: Site name, description, timezone, language

### User Experience Design

#### **Navigation Flow**
1. **Authentication** → Login/Register with proper error handling
2. **Dashboard** → Central hub with overview and quick actions
3. **Forms** → Comprehensive form management with intuitive workflows
4. **Submissions** → Efficient data review and management
5. **Settings** → Organized configuration with clear sections

#### **Responsive Design**
- **Desktop** (1024px+): Full sidebar navigation, multi-column layouts
- **Tablet** (768px-1024px): Collapsible navigation, stacked layouts
- **Mobile** (320px-768px): Bottom navigation, single-column flows

#### **Accessibility Features**
- **Keyboard Navigation**: Full keyboard accessibility for all functions
- **Screen Reader Support**: Proper ARIA labels and semantic HTML
- **Color Contrast**: WCAG AA compliant color schemes
- **Focus Indicators**: Clear visual focus states for all interactive elements
- **Error Handling**: Descriptive error messages with suggested actions

<!-- IGNORE FOR NOW -->
<!-- #### **Performance Considerations**
- **Lazy Loading**: Forms and submissions loaded on demand
- **Caching**: Template and data caching for faster page loads
- **Compression**: Optimized assets and response compression
- **Progressive Enhancement**: Core functionality works without JavaScript -->

### Technical Implementation

#### **Template Architecture**
- **Layouts**: Base HTML structure with blocks for content injection
- **Pages**: Content templates that extend layouts
- **Components**: Reusable UI elements with parameterized data
- **Atoms**: Basic building blocks for consistent styling

#### **Data Flow**
1. **Request Routing** → URL-based page selection
2. **Authentication** → Session validation and user context
3. **Data Loading** → Database queries and data preparation
4. **Template Rendering** → Server-side HTML generation with Go templates
5. **Response** → Optimized HTML delivery to client

#### **Security Features**
- **CSRF Protection**: All forms protected against cross-site request forgery
- **Input Validation**: Server-side validation using go-playground/validator
- **Session Security**: Secure session management with expiration
- **SQL Injection Prevention**: Parameterized queries only
- **XSS Protection**: Automatic template escaping and content sanitization

### Color Schemes & Themes

#### **Robot Green** (Default)
- **Primary**: Vibrant green (#00ff88)
- **Secondary**: Soft teal (#00d4aa)
- **Accent**: Electric blue (#0099ff)
- **Character**: Modern, tech-focused, high-energy

#### **Midnight Wisp**
- **Primary**: Deep purple (#6b46c1)
- **Secondary**: Muted lavender (#8b5cf6)
- **Accent**: Soft gold (#fbbf24)
- **Character**: Professional, elegant, sophisticated

#### **Pale Wisp**
- **Primary**: Soft blue (#3b82f6)
- **Secondary**: Light gray (#6b7280)
- **Accent**: Warm orange (#f59e0b)
- **Character**: Clean, minimalist, approachable

### Development Guidelines

#### **Template Best Practices**
- **No HTML comments with template syntax** - Causes recursion issues
- **Use atomic components** - Build complex UIs from simple, reusable parts
- **Consistent naming** - Follow established patterns for template names
- **Error handling** - Always provide fallback states and error messages
- **Data validation** - Validate all inputs server-side with proper error display

#### **Component Usage**
```go
// Correct: Use template calls without HTML comments
{{template "atoms/button" dict "text" "Save" "style" "btn-primary"}}

// Correct: Pass data through dict helper
{{template "components/form-field" dict "label" "Email" "type" "email" "name" "email" "required" true}}

// Correct: Handle conditional rendering
{{if .hasError}}
    {{template "atoms/alert" dict "type" "alert-error" "message" .errorMessage}}
{{end}}
```

#### **Form Handling Standards**
- **All inputs use `application/x-www-form-urlencoded`** - No JSON input
- **Server-side validation** - Using go-playground/validator
- **CSRF protection** - All forms must include CSRF tokens
- **Error states** - Visual feedback for validation failures
- **Success feedback** - Clear confirmation of successful actions

### API Standards

#### **Request Format**
- **Form Data**: All input as `application/x-www-form-urlencoded`
- **Query Parameters**: For filtering, pagination, and navigation
- **Headers**: Authentication and debug flags

#### **Response Format**
- **HTML Pages**: Server-rendered templates for main interface
- **Error Responses**: Plain text with optional debug information
- **Debug Mode**: Include debug info when `__include_debug_info__=true`

#### **Error Handling**
- **User-friendly messages** - Clear, actionable error descriptions
- **Proper HTTP status codes** - 400 for validation, 401 for auth, etc.
- **Logging** - All errors logged for debugging and monitoring
- **Graceful degradation** - Fallback states for system failures

### Future Enhancements

#### **Planned Features**
- **Multi-site Management** - Complete project/site management interface
- **Advanced Analytics** - Detailed form performance and user behavior metrics
- **Form Builder** - Visual drag-and-drop form creation interface
- **Webhook Integration** - Real-time submission notifications and data sync
- **API Endpoints** - RESTful API for external integrations
- **Plugin System** - Extensible architecture for custom functionality
- **Advanced Theming** - Custom CSS injection and theme marketplace
- **Collaboration Tools** - Multi-user editing and approval workflows

#### **Technical Roadmap**
- **Database Migrations** - Automated schema updates and version management
- **Caching Layer** - Redis integration for improved performance
- **Search Integration** - Full-text search across all content
- **Backup System** - Automated backups with point-in-time recovery
- **Monitoring Dashboard** - System health and performance metrics
- **Security Audit** - Regular security assessments and penetration testing

This design document serves as the foundation for Wispy CMS development, ensuring consistency, usability, and maintainability across all features and interfaces.
