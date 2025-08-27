package util

func OptionalString(str string) *string {
	if str != "" {
		return &str
	}
	return nil
}
