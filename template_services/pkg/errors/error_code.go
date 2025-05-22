package errors

import "log"

func GetAppErrorMessage(code string) string {

	statusMessages := map[string]string{
		// TMPxxx (Template Service)
		TmpErrInvalidRequestBody:        "Invalid request body",
		TmpErrUUIDParsing:               "Invalid UUID in URL",
		TmpErrTemplateCreate:            "Failed to create template",
		TmpErrTemplateFetch:             "Failed to get template",
		TmpErrTemplateUpdate:            "Failed to update template",
		TmpErrTemplateDelete:            "Template not found",
		TmpErrTemplateNotFound:          "Failed to delete template",
		TmpErrTemplateIDNotFound:        "Template ID not found",
		TmpErrmissingName:               "Please provide name",
		TmpErrmissingChannel:            "Please provide channel",
		TmpErrmissingLanguage:           "Please provide language",
		TmpErrmissingContent:            "Please provide content",
		TmpErrmissingIsActive:           "Please provide is_active status",
		TmpErrmissingTemplateID:         "Please provide template ID",
		TmpErrTemplateNameAlreadyExists: "Template name already exists",
		TmpErrTemplateList:              "Failed to retrieve templates",
		TmpErrmissingRequiredFields:     "Please provide required fields",
		TmpErrTemplateListEmpty:         "No templates found",
		// Add other groups (JWT, etc.) here...
	}

	if message, exists := statusMessages[code]; exists {
		log.Printf("Error code: %s, Message: %s", code, message)
		return message
	}
	return "An unknown error occurred"
}
