package validation

import (
	"net"
	"regexp"
	"strings"
)

func isEmailSyntaxValid(email string) bool {
	// checking for email syntax
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	match, _ := regexp.MatchString(regex, email)
	return match
}

func isDomainValid(email string) bool {
	// checking for valid mx records for the domain
	parts := strings.Split(email, "@")

	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	_, err := net.LookupMX(domain)

	return err == nil
}
