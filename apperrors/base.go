package apperrors

type AppError interface {
	error
	HTTPCode() int
	HTTPResponse() map[string]any
}

type baseError struct {
	Err      error
	Resource string
}

func (e baseError) Error() string { return e.Err.Error() }

func (e baseError) httpResponse(title string) map[string]any {
	return map[string]any{
		"title":  title, // too brittle here, should be defined for every error
		"detail": e.Error(),
		"entity": e.Resource,
	}
}
