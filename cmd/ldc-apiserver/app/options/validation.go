package options

// Validate validates server run options, to find
// options' misconfiguration
func (s *ServerRunOptions) Validate() []error {
	var errors []error

	errors = append(errors, s.GenericServerRunOptions.Validate()...)
	// errors = append(errors, s.RedisOptions.Validate()...)
	return errors
}
