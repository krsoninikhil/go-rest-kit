package apperrors

type AppError interface {
	error
	HTTPCode() int
	HTTPResponse() map[string]any
}

type baseError struct {
	Cause    error
	Resource string
}

func (e baseError) Error() string { return e.Cause.Error() }

func (e baseError) httpResponse(title string) map[string]any {
	return map[string]any{
		"title":  title,
		"detail": e.Error(),
		"entity": e.Resource,
	}
}
