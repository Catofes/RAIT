package misc

// OrDefault returns the given string, or the default value if it is empty
func OrDefault(value string, def string) string {
	if value == "" {
		return def
	}
	return value
}
