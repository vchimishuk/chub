package cue

// stringTruncate truncates string up to newLen characters.
// If given string is shorter than newLen if will be returned without any changes.
func stringTruncate(str string, newLen int) string {
	if len(str) > newLen {
		return str[:newLen]
	}

	return str
}
