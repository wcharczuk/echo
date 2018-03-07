package protoutil

// Error is an error.
type Error string

// Error implements error
func (e Error) Error() string {
	return string(e)
}
