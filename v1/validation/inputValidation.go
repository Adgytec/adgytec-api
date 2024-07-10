package validation

import "regexp"

const (
	superAdmin string = "super_admin"
	admin      string = "admin"
	user       string = "user"
)

func ValidateEmail(email string) bool {
	// validating email syntax and checking for valid email domain
	return isEmailSyntaxValid(email) && isDomainValid(email)
}

func ValidateRole(role string) bool {
	// checking if the new role is a valid role
	return role == superAdmin || role == admin || role == user
}

func AuthorizeRole(myRole, role string) bool {
	if role == user {
		return true
	}

	return myRole == superAdmin
}

func ValidateName(name string) bool {
	regex := `^[a-zA-Z]+(?:[' -][a-zA-Z]+)*$`
	match, _ := regexp.MatchString(regex, name)
	// checking if name is empty or not
	return match && len(name) >= 3
}

func ValidatePhone(phone string) bool {
	regex := `^(\+?\d{1,3})?[- .]?\d{3}[- .]?\d{3}[- .]?\d{4}$`

	match, _ := regexp.MatchString(regex, phone)
	return match
}
