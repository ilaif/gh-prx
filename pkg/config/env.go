package config

import (
	"os"
)

func GetOpenAIApiKey() string {
	return os.Getenv("OPENAI_API_KEY")
}
