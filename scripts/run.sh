#!/bin/bash

# Navigate to project root directory (parent of scripts folder)
cd "$(dirname "$0")/.." || { echo "Failed to navigate to project root"; exit 1; }
PROJECT_ROOT=$(pwd)
RED='\033[0;31m'
DARK_GRAY='\033[1;30m'
RESET='\033[0m'
# Kill any process running on port 8080
echo "${RED}Shutting down any running server on port 8080...  ${RESET}"
lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No server was running"
go mod tidy

# # Create .env file if it doesn't exist
# if [ ! -f .env ]; then
#   echo "${RED} No .env found! ${DARK _GRAY}Creating default .env file... ${RESET}"
#   cat > .env << EOL
# # .env file for Wispy Core CMS
# PORT=8080
# HOST=localhost
# ENV=development

# # Wispy Path
# CACHE_DIR=.wispy
# SITES_PATH=_data/tenants
# STATIC_PATH=_data/design/static
# GLOBAL_TEMPLATES=_data/design/templates

# # Site Paths
# MIGRATION_ROOT=migrations
# DATABASE_PATH=dbs/migrations


# # Api/Security
# RATE_LIMIT_REQUESTS_PER_SECOND=12
# RATE_LIMIT_REQUESTS_PER_MINUTE=240

# # AUTH
# DISCORD_CLIENT_ID=
# DISCORD_CLIENT_SECRET=
# DISCORD_REDIRECT_URI=

# # Wispy Debugging
# DEBUG_WISPY_TAIL=false


# EOL
#   echo "${DARK_GRAY} .env file created ${RESET}"
# fi

# Load environment variables from .env
# if [ -f .env ]; then
#   # echo "Loading environment variables from .env file"
#   export $(grep -v '^#' .env | xargs)
# fi

# Set environment variables if not present
# export PORT=${PORT:-8080}
# export WISPY_CORE_ROOT=${PROJECT_ROOT}
# export SITES_PATH=${SITES_PATH:-data/sites}

# Ensure sites directory exists
# if [[ "${SITES_PATH}" == /* ]]; then
#   # Absolute path
#   SITES_DIR="${SITES_PATH}"
# else
#   # Relative path
#   SITES_DIR="${PROJECT_ROOT}/${SITES_PATH}"
# fi

# echo "Using sites directory: ${SITES_DIR}"
# if [ ! -d "$SITES_DIR" ]; then
#   # echo "Creating sites directory at ${SITES_DIR}"
#   mkdir -p "$SITES_DIR"
  
#   # Create a default localhost site structure if it doesn't exist
#   LOCALHOST_DIR="${SITES_DIR}/localhost"
#   if [ ! -d "$LOCALHOST_DIR" ]; then
#     echo "Creating a default localhost site structure"
#     mkdir -p "${LOCALHOST_DIR}/pages"
#     mkdir -p "${LOCALHOST_DIR}/layouts"
#     mkdir -p "${LOCALHOST_DIR}/templates/partials"
#     mkdir -p "${LOCALHOST_DIR}/templates/sections"
#     mkdir -p "${LOCALHOST_DIR}/public/css"
#     mkdir -p "${LOCALHOST_DIR}/public/js"
#     mkdir -p "${LOCALHOST_DIR}/assets/css"
#     mkdir -p "${LOCALHOST_DIR}/assets/js"
#     mkdir -p "${LOCALHOST_DIR}/config"
#     mkdir -p "${LOCALHOST_DIR}/dbs"
    
#     # Create a minimal page
#     cat > "${LOCALHOST_DIR}/pages/home.html" << EOL
# <h1>Welcome to Wispy Core</h1>
# <p>This is a default page created by the installation script.</p>
# EOL

#     # Create a minimal layout
#     cat > "${LOCALHOST_DIR}/layouts/default.html" << EOL
# <!DOCTYPE html>
# <html>
# <head>
#   <title>{{ .Title }}</title>
#   <meta charset="utf-8">
#   <meta name="viewport" content="width=device-width, initial-scale=1">
# </head>
# <body>
#   <!-- Main Content -->
#   <main class="p-6">
#       {% block "page-body" %}<p> Could not load page body.</p> {% endblock %}
#   </main>
#   <!-- Footer -->
#   {% block "footer" %}
#       <footer class="footer footer-center p-4 text-base-content">
#           <div>
#               <p>&copy; 2025 {{ Site.Name }}. Powered by Wispy-CMS.</p>
#           </div>
#       </footer>
#   {% endblock %}
# </body>
# </html>
# EOL
#   fi
# fi
# export HOST=${HOST:-localhost}
# export ENV=${ENV:-development}
# export SITES_PATH=${SITES_PATH:-${PROJECT_ROOT}/sites}

# # Ensure the sites directory exists
# # echo "Ensuring sites directory exists at: $SITES_PATH"
# mkdir -p "$SITES_PATH"

# Display startup information
# echo "Starting Wispy Core CMS with settings:"
# echo "  - Host: $HOST"
# echo "  - Port: $PORT"
# echo "  - Environment: $ENV"
# echo "  - Sites path: $SITES_PATH"

# Run the server from the cmd/server directory
# echo "Starting server from project root: ${PROJECT_ROOT}"
go run ./server/server.go

# Exit with the same status code as the server
exit $?
