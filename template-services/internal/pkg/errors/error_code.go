package appErrors

func GetAppErrorMessage(code string) string {

	statusMessages := map[string]string{
		// TMPxxx (Template Service)
		"TmpErrInvalidRequestBody":        "Invalid request body",
		"TmpErrUUIDParsing":               "Invalid UUID in URL",
		"TmpErrTemplateCreate":            "Failed to create template",
		"TmpErrTemplateFetch":             "Failed to get template",
		"TmpErrTemplateUpdate":            "Failed to update template",
		"TmpErrTemplateNotFound":          "Template not found",
		"TmpErrTemplateDelete":            "Failed to delete template",
		"TmpErrmissingName":               "Please provide name",
		"TmpErrmissingChannel":            "Please provide channel",
		"TmpErrmissingLanguage":           "Please provide language",
		"TmpErrmissingContent":            "Please provide content",
		"TmpErrmissingIsActive":           "Please provide is_active status",
		"TmpErrmissingTemplateID":         "Please provide template ID",
		"TmpErrTemplateIDNotFound":        "Template ID not found",
		"TmpErrTemplateNameAlreadyExists": "Template name already exists",
		// Add other groups (JWT, etc.) here...
	}

	if message, exists := statusMessages[code]; exists {
		return message
	}
	return "An unknown error occurred"
}
