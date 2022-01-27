package providers

type ProviderErrorType string

const (
	ProviderErrorNotFound    ProviderErrorType = "NOT_FOUND"
	ProviderErrorConflict    ProviderErrorType = "CONFLICT"
	ProviderErrorForbidden   ProviderErrorType = "FORBIDDEN"
	ProviderErrorServerError ProviderErrorType = "SERVER_ERROR"
)
