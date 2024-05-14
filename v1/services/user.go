package services

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"

	"github.com/jackc/pgx/v5"
	"github.com/rohan031/adgytec-api/firebase"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/validation"
)

type UserCreationDetails struct {
	Name     string
	Email    string
	Password string
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func generateRandomPassword() (string, error) {
	// Define character sets for password generation
	upperChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars := "abcdefghijklmnopqrstuvwxyz"
	digitChars := "0123456789"
	specialChars := "!@#$%^&*()-_=+[]{}|;:,.<>?~"

	// Concatenate all character sets
	allChars := upperChars + lowerChars + digitChars + specialChars

	var password strings.Builder
	for i := 0; i < 10; i++ {
		// Generate random index to select a character from allChars
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", err
		}

		// Append selected character to password
		password.WriteByte(allChars[randomIndex.Int64()])
	}

	return password.String(), nil
}

func (u *User) CreateUser() (string, error) {
	// creating random password
	password, err := generateRandomPassword()
	if err != nil {
		log.Printf("Error generating password: %v\n", err)
		return "", err
	}

	// creating user in firebase
	params := (&auth.UserToCreate{}).Email(u.Email).DisplayName(u.Name).Password(password)
	userRecord, err := firebase.FirebaseClient.CreateUser(context.Background(), params)
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			message := "The email address provided is already associated with an existing user account."
			return "", &custom.MalformedRequest{Status: http.StatusConflict, Message: message}
		}

		log.Printf("Error creating user in firebase: %v\n", err)
		return "", err
	}

	// setting custom claims for newly created user
	uid := userRecord.UID
	claims := map[string]interface{}{"role": u.Role}
	err = firebase.FirebaseClient.SetCustomUserClaims(context.Background(), uid, claims)
	if err != nil {
		log.Printf("Error setting custom claims: %v\n", err)
		return "", err
	}

	// inserting into database user table
	query := `INSERT INTO users (user_id, name, email, role) values (@userId, @name, @email, @role	)`
	args := pgx.NamedArgs{
		"userId": uid,
		"email":  u.Email,
		"name":   u.Name,
		"role":   u.Role,
	}

	_, err = db.Exec(context.Background(), query, args)
	if err != nil {
		log.Printf("Error adding user in database: %v\n", err)
		return "", err
	}

	return password, nil
}

func (u *User) ValidateInput() bool {
	// validating email, role and name parameters
	return (validation.ValidateEmail(u.Email) &&
		validation.ValidateRole(u.Role) &&
		validation.ValidateName(u.Name))
}
