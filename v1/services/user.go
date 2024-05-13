package services

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5"
	"github.com/rohan031/adgytec-api/firebase"
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
	// use admin sdk to create user in firebase
	// save newly created user to user table
	password, err := generateRandomPassword()
	if err != nil {
		log.Printf("Error generating password: %v\n", err)
		return "", err
	}

	params := (&auth.UserToCreate{}).Email(u.Email).DisplayName(u.Name).Password(password)
	userRecord, err := firebase.FirebaseClient.CreateUser(context.Background(), params)
	if err != nil {
		log.Printf("Error creating user in firebase: %v\n", err)
		return "", err
	}

	uid := userRecord.UID
	claims := map[string]interface{}{"role": u.Role}
	err = firebase.FirebaseClient.SetCustomUserClaims(context.Background(), uid, claims)
	if err != nil {
		log.Printf("Error setting custom claims: %v\n", err)
		return "", err
	}

	// inserting into database
	query := `INSERT INTO users (user_id, name, email) values (@userId, @name, @email)`
	args := pgx.NamedArgs{
		"userId": uid,
		"email":  u.Email,
		"name":   u.Name,
	}

	_, err = db.Exec(context.Background(), query, args)
	if err != nil {
		log.Printf("Error adding user in database: %v\n", err)
		return "", err
	}

	return password, nil
}

func (u *User) ValidateInput() bool {
	return validation.ValidateEmail(u.Email) && validation.ValidateRole(u.Role)
}
