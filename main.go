package main

import (
	"os"

	"github.com/joho/godotenv"

	"wispy-core/common"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		common.Fatal("Error loading .env file: %v", err)
	}

	// get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}
	// Set the WISPY_CORE_ROOT environment variable to the current directory
	os.Setenv("WISPY_CORE_ROOT", currentDir)
}
