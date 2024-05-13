package validation

func ValidateEmail(email string) bool {
	return isEmailSyntaxValid(email) && isDomainValid(email)
}

func ValidateRole(role string) bool {
	const (
		superAdmin string = "super_admin"
		admin      string = "admin"
		user       string = "user"
	)

	return role == superAdmin || role == admin || role == user
}
