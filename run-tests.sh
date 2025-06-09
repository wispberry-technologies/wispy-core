# lsof -ti:8080 | xargs kill -9  
export TEST_MODE=true && go run main.go template
