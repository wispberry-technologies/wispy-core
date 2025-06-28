# Kill any process running on port 8080
echo "${RED}Shutting down any running server on port 8080...  ${RESET}"
lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No server was running"