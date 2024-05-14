package validation

func ValidateEmail(email string) bool {
	// validating email syntax and checking for valid email domain
	return isEmailSyntaxValid(email) && isDomainValid(email)
}

func ValidateRole(role string) bool {
	// checking if the new role is a valid role
	const (
		superAdmin string = "super_admin"
		admin      string = "admin"
		user       string = "user"
	)

	return role == superAdmin || role == admin || role == user
}

func ValidateName(name string) bool {
	// checking if name is empty or not
	return len(name) >= 3
}
