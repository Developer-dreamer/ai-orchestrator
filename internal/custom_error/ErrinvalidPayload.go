package custom_error

type InvalidPayloadError struct {
	Message string
}

func (e *InvalidPayloadError) Error() string {
	return e.Message
}

func NewErrInvalidPayload(message string) error {
	return &InvalidPayloadError{message}
}
