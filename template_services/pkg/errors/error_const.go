package errors

const (
	// TMPxxx (Template Service Errors)

	TmpErrInvalidRequestBody        = "TMP001"
	TmpErrUUIDParsing               = "TMP002"
	TmpErrTemplateCreate            = "TMP003"
	TmpErrTemplateFetch             = "TMP004"
	TmpErrTemplateUpdate            = "TMP005"
	TmpErrTemplateDelete            = "TMP006"
	TmpErrTemplateNotFound          = "TMP007"
	TmpErrTemplateIDNotFound        = "TMP008"
	TmpErrmissingName               = "TMP009"
	TmpErrmissingChannel            = "TMP0010"
	TmpErrmissingLanguage           = "TMP0011"
	TmpErrmissingContent            = "TMP0012"
	TmpErrmissingIsActive           = "TMP0013"
	TmpErrmissingTemplateID         = "TMP0014"
	TmpErrTemplateNameAlreadyExists = "TMP0015"
	TmpErrTemplateList              = "TMP0016"
	TmpErrmissingRequiredFields     = "TMP0017"
	TmpErrTemplateListEmpty         = "TMP0018"
)
