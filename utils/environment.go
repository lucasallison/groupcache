package utils

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
	env, ok := os.LookupEnv("environment")
	if !ok {
		return ".env.production"
	}
	return env
}

func GetHostFromEnv() string {
	LoadEnv()
	val, ok := os.LookupEnv("HOST")
	if !ok {
		// TODO error?
		return ""
	}
	return val
}

func PrefetchingEnabled() bool {
	LoadEnv()
	val, ok := os.LookupEnv("PREFETCH")
	if !ok {
		return false
	}
	return val == "true"
}

func LookUpEnvVariable(key string) (val string, ok bool) {
	LoadEnv()
	return os.LookupEnv(key)
}
