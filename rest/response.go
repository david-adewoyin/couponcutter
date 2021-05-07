package rest

//ResponseError represent the error message sent to the client
type ResponseError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Desc    string `json:"desc,omitempty"`
}

//ResponseErrorWithField represent the error message sent to the client with the field  in which the error occurred
type ResponseErrorWithField struct {
	Code    int    `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
	Desc    string `json:"desc,omitempty"`
}

//A helper for constructing responseErrors
func constructError(code int, message string, desc string) ResponseError {

	return ResponseError{
		Code:    code,
		Message: message,
		Desc:    desc,
	}

}

//A helper for constructing responseErrors with the fields in which it occurs
func constructErrorWithField(code int, field string, message string, desc string) ResponseErrorWithField {
	return ResponseErrorWithField{
		Code:    code,
		Field:   field,
		Message: message,
		Desc:    desc,
	}

}
