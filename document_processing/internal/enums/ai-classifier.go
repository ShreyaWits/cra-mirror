//go:build !test
// build +test
package enums

type AIClassifier string

const (
	ClassifierLlama  AIClassifier = "llama"
	ClassifierGemini AIClassifier = "gemini"
)