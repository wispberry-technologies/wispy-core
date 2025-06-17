# Wispy Core Templates - DaisyUI Implementation

This directory contains the core application templates built with DaisyUI 5 standards for the Wispy framework.

## Overview

All templates follow DaisyUI 5 best practices and include:
- Modern, responsive design using DaisyUI components
- Proper form controls with validation
- AJAX-based authentication
- OAuth provider integration (Discord only)
- Reusable partials for common functionality
- Consistent color schemes and theming support

## Templates Structure

```
templates/app/
├── account-login.html      # Login page with OAuth support
├── account-register.html   # Registration page
├── dashboard.html          # Main dashboard (demo)
├── settings.html          # Settings page (demo)
└── partials/
    ├── auth-guard.html     # Authentication protection
    ├── user-nav.html       # User navigation dropdown
    ├── password-change.html # Password change form
    └── user-profile.html   # User profile management
```

## Authentication Pages

### account-login.html
- Hero layout with form and welcome message
- Email/password login with validation
- OAuth provider button (Discord)
- Loading states and error handling
- AJAX form submission to `/api/v1/login`

### account-register.html
- Clean registration form with fieldset
- Form validation (password confirmation, requirements)
- OAuth provider options
- Success/error messaging
- AJAX form submission to `/api/v1/register`

## Reusable Partials

### auth-guard.html
Authentication protection script that:
- Checks user authentication status
- Shows loading indicator during check
- Redirects to login if not authenticated
- Supports role-based access control
- Provides utility functions for auth state

**Usage:**
```html
{% include "templates/app/partials/auth-guard.html" %}
```

**Configuration:**
```javascript
const AUTH_CONFIG = {
  redirectTo: '/login',    // Login redirect URL
  checkOnLoad: true,       // Check auth on page load
  requiredRoles: [],       // Required user roles (empty = any authenticated user)
  showLoading: true        // Show loading indicator
};
```

### user-nav.html
User navigation dropdown that:
- Shows loading spinner during auth check
- Displays user info and avatar when authenticated
- Shows login/register buttons when not authenticated
- Includes logout functionality
- Responsive design with icons

**Usage:**
```html
<!-- Include in navbar-end -->
{% include "templates/app/partials/user-nav.html" %}
```

### password-change.html
Password change form component:
- Current password verification
- New password with confirmation
- Password requirements display
- AJAX form submission to `/api/v1/change-password`
- Success/error messaging

**Usage:**
```html
{% include "templates/app/partials/password-change.html" %}
```

### user-profile.html
User profile management form:
- Profile picture upload (placeholder)
- Personal information editing
- Account status badges (email verified, account status, 2FA)
- Form validation and AJAX submission
- Auto-loads current user data

**Usage:**
```html
{% include "templates/app/partials/user-profile.html" %}
```

## Demo Pages

### dashboard.html
Complete dashboard demonstration showing:
- Authentication guard implementation
- User navigation integration
- Project statistics cards
- Recent projects list
- Quick actions panel
- Modal for creating new projects
- Integration of profile and password change components

### settings.html
Settings page with tabbed interface:
- **Profile Tab**: User profile component
- **Security Tab**: Password change and 2FA settings
- **Preferences Tab**: Theme selection and notification settings
- **Danger Zone Tab**: Account deletion with confirmation

## DaisyUI Components Used

### Forms
- `form-control` - Form field wrapper
- `label` - Field labels with proper structure
- `input` - Text inputs with borders and validation states
- `textarea` - Multi-line text inputs
- `select` - Dropdown selectors
- `btn` - Buttons with various styles and states

### Layout
- `hero` - Hero sections with content
- `card` - Content cards with headers and actions
- `navbar` - Navigation bars
- `container` - Content containers
- `grid` - Responsive grid layouts

### Feedback
- `alert` - Success/error/info messages
- `loading` - Loading spinners
- `badge` - Status indicators
- `stats` - Statistics displays

### Navigation
- `dropdown` - User dropdown menus
- `menu` - Navigation menus
- `tabs` - Tabbed interfaces
- `breadcrumbs` - Navigation breadcrumbs

### Interactive
- `modal` - Dialog boxes
- `toggle` - Toggle switches
- `checkbox` - Checkboxes
- `radio` - Radio buttons

## API Integration

All forms integrate with the authentication API:

- **POST** `/api/v1/login` - User login
- **POST** `/api/v1/register` - User registration  
- **POST** `/api/v1/logout` - User logout
- **GET** `/api/v1/me` - Get current user info
- **POST** `/api/v1/change-password` - Change password
- **GET** `/api/v1/oauth/{provider}` - OAuth login
- **GET** `/api/v1/projects` - Get user projects
- **POST** `/api/v1/projects` - Create new project

## Theme Support

Templates support DaisyUI themes through:
- `data-theme` attribute on `<html>` element
- Theme controller inputs for theme switching
- Semantic color classes (`primary`, `secondary`, etc.)
- Automatic dark/light mode handling

## JavaScript Features

### AJAX Forms
All forms use modern fetch API with:
- Proper error handling
- Loading states
- Success/error messaging
- Form validation
- Credential inclusion for cookies

### User State Management
- Global `currentUser` object
- Authentication status checking
- Role-based access helpers
- Session state management

### UI Interactions
- Modal management
- Theme switching
- Responsive navigation
- Toast notifications
- Dynamic content loading

## Responsive Design

All templates are fully responsive:
- Mobile-first approach
- Breakpoint-aware layouts (`sm:`, `lg:`, etc.)
- Responsive navigation (hamburger menu)
- Flexible grid systems
- Touch-friendly interactions

## Accessibility

Templates include accessibility features:
- Proper ARIA labels
- Screen reader support
- Keyboard navigation
- Focus management
- Semantic HTML structure
- Color contrast compliance

## Customization

### Styling
- Use DaisyUI utility classes for customization
- Override with Tailwind utilities when needed
- Custom CSS should be minimal
- Follow DaisyUI design tokens

### Components
- Partials are modular and reusable
- Easy to customize via template variables
- JavaScript functions are standalone
- Configuration objects for flexibility

## Best Practices

1. **Forms**: Always use `form-control` wrapper for form fields
2. **Colors**: Use semantic DaisyUI colors (`primary`, `error`, etc.)
3. **Loading**: Show loading states for async operations
4. **Validation**: Client-side validation with server-side backup
5. **Errors**: Clear, user-friendly error messages
6. **Responsive**: Design mobile-first, enhance for larger screens
7. **Accessibility**: Include ARIA labels and semantic markup

## Browser Support

Templates work in all modern browsers:
- Chrome 90+
- Firefox 90+
- Safari 14+
- Edge 90+

## Dependencies

- DaisyUI 5.x
- Tailwind CSS 4.x
- Modern browsers with fetch API support
- No additional JavaScript libraries required

## Integration with Wispy

Templates integrate seamlessly with:
- Wispy's template engine
- Authentication system
- Dynamic API routing
- Site-specific theming
- Database models

For implementation details, see the main Wispy documentation.
