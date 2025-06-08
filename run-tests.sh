lsof -ti:8080 | xargs kill -9  

export TEST_MODE=false && go run main.go
