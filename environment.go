package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	err = godotenv.Load(filepath.Join(workingDir, RetrieveEnvFile()))
	if err != nil {
		log.Println(err)
	}
}

func RetrieveEnvFile() string {
	env, ok := os.LookupEnv("EnvFile")
	if !ok {
		return ".env.local"
	}
	return env
}

func GetHostFromEnv() string {
	LoadEnv()
	env, ok := os.LookupEnv("HOST")
	if !ok {
		// TODO error?
		return ""
	}
	return env
}
