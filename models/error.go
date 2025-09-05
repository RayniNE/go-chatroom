package models

type CustomError struct {
	Code       int
	Message    string
	AppContext string
}

func (e *CustomError) Error() string {
	return e.Message
}

type ServerResponse struct {
	Data any
	Code int
}
