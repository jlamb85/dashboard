# Changelog

All notable changes to the Server Dashboard project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Authentication System**: Comprehensive user authentication with bcrypt password hashing
  - Login/logout functionality with session management
  - Secure session cookies with HMAC signing
  - Multi-user support with configurable users in config.yaml
  - Password change UI for authenticated users
  - Admin-only user creation interface at `/account/users/new`
  
- **Group-Based Authorization**: Flexible permission management system
  - Groups configuration in auth.groups with name, description, and permissions
  - Users can belong to multiple groups (comma-separated)
  - Admin group requirement for creating users and managing groups
  - Manage Groups UI at `/account/groups` for defining and editing groups
  - Backward compatibility with legacy `roles` field
  
- **Session Middleware**: Automatic session validation and authentication enforcement
  - 8-hour session duration
  - Automatic redirect to login for unauthenticated requests
  - Public endpoint exceptions (login, logout, health, static assets)
  
- **Synthetic Monitoring**: External service health checks
  - HTTP, TCP, and DNS check types
  - Configurable intervals and timeouts
  - Status visualization with success/failure tracking
  - `/synthetics` page with real-time check results
  
- **UI Navigation Enhancements**:
  - "Change Password" links in all page navbars and sidebars
  - "Create User" links (admin-only) in all page navbars and sidebars
  - "Manage Groups" links (admin-only) in all page navbars and sidebars
  - "Logout" links in all page navbars and sidebars
  - Theme-safe navigation link styling with `.nav-link-utility` CSS class

### Changed
- **Config Structure**: Extended authentication configuration
  - Added `auth.users[]` array for multi-user mode
  - Added `auth.session_secret` for session cookie signing
  - Added `auth.groups[]` for group definitions with permissions
  - Users can have both `groups` and legacy `roles` fields
  
- **User Management**: All password handling now uses bcrypt
  - Plaintext passwords automatically removed when hash is set
  - Config persistence with in-memory updates
  - User creation appends to config file with proper YAML formatting

### Security
- **Password Security**: Bcrypt hashing for all user passwords
  - Default cost factor (12 rounds)
  - Plaintext password fallback only for development
  - Tools/hashpass utility for generating secure hashes
  
- **Session Security**: Cryptographically signed session cookies
  - HMAC-SHA256 signing with configurable secret
  - HttpOnly cookies to prevent XSS attacks
  - Expiration validation on every request

### Fixed
- Light theme navbar link visibility (added contrast styling)
- Template name consistency (lowercase HTML extensions)
- Monitoring page layout and controls positioning

### Technical Details
- **New Files**:
  - `internal/handlers/auth.go`: Login/logout handlers
  - `internal/handlers/password.go`: Password change handler
  - `internal/handlers/user_admin.go`: Admin user creation handler
  - `internal/handlers/groups.go`: Groups management handler
  - `internal/handlers/synthetics.go`: Synthetic checks handler
  - `internal/middleware/session.go`: Session management middleware
  - `web/templates/login.html`: Login page
  - `web/templates/change-password.html`: Password change form
  - `web/templates/new-user.html`: User creation form
  - `web/templates/groups.html`: Groups management interface
  - `web/templates/monitoring.html`: Monitoring page
  
- **Routes Added**:
  - `GET/POST /login`: User authentication
  - `GET/POST /logout`: Session termination
  - `GET/POST /account/password`: Password change
  - `GET/POST /account/users/new`: Admin user creation
  - `GET/POST /account/groups`: Groups management
  - `GET /synthetics`: Synthetic monitoring view

## [1.0.0] - 2026-01-24

### Initial Release
- Server monitoring dashboard
- Virtual machine tracking
- Switch monitoring
- Mock data mode for development
- Bootstrap-based responsive UI
- Dark/light theme toggle
