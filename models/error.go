package models

type CustomError struct {
	Code    int
	Message string
}

type ServerResponse struct {
	Data any
	Code int
}
