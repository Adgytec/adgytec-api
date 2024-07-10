package services

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"firebase.google.com/go/v4/auth"

	"github.com/jackc/pgx/v5"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
	"github.com/rohan031/adgytec-api/v1/validation"
)

type UserCreationDetails struct {
	Name     string
	Email    string
	Password string
}

type User struct {
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`
	UserId    string    `json:"userId,omitempty" db:"user_id"`
	CreatedAt time.Time `json:"createdAt,omitempty" db:"created_at"`
	Cursor    int       `json:"-" db:"cursor"`
}

/*
create user
*/
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

/*
if user exists in db return true
else delete user from firebase and re-add it
return false, err // internal server error
return true, nil // user exits
return false, nil // delete user from firebase and create new user
*/
func userExistsInDb(email string) (bool, error) {
	// fetching the user from db
	args := dbqueries.GetUserByEmailArgs(email)
	rows, err := db.Query(ctx, dbqueries.GetUserByEmail, args)
	if err != nil {
		log.Printf("Error fetching user from db: %v\n", err)
		return false, err
	}
	defer rows.Close()

	_, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// user doesn't exist in db
			u, err := firebaseClient.GetUserByEmail(ctx, email)
			if err != nil {
				log.Printf("Error getting user data from firebase: %v\n", err)
				return false, err
			}

			err = firebaseClient.DeleteUser(ctx, u.UID)
			if err != nil {
				log.Printf("Error deleting user from firebase: %v\n", err)
				return false, err
			}

			return false, nil
		}
		log.Printf("Error reading rows: %v\n", err)
		return false, err
	}

	return true, nil
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
	userRecord, err := firebaseClient.CreateUser(ctx, params)
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			// find user in db
			ispresent, err := userExistsInDb(u.Email)
			if err != nil {
				return "", err
			}

			// user already exists
			if ispresent {
				message := "The email address provided is already associated with an existing user account."
				return "", &custom.MalformedRequest{Status: http.StatusConflict, Message: message}
			}

			// create new user with given details
			return u.CreateUser()
		}

		log.Printf("Error creating user in firebase: %v\n", err)
		return "", err
	}

	// setting custom claims for newly created user
	uid := userRecord.UID
	claims := map[string]interface{}{"role": u.Role}
	err = firebaseClient.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		log.Printf("Error setting custom claims: %v\n", err)
		return "", err
	}

	// inserting into database user table
	args := dbqueries.CreateUserArgs(uid, u.Email, u.Name, u.Role)
	_, err = db.Exec(ctx, dbqueries.CreateUser, args)
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

/*
update user:
updateUser() => updates user role and name
updateUserName() => updates user name
*/
func (u *User) ValidateUpdateInput() bool {
	// validating email, role and name parameters
	if len(u.Role) == 0 && len(u.Name) == 0 {
		return false
	}

	if (len(u.Role) > 0 && !validation.ValidateRole(u.Role)) ||
		(len(u.Name) > 0 && !validation.ValidateName(u.Name)) {
		return false
	}

	return true
	// return (validation.ValidateRole(u.Role) ||
	// 	validation.ValidateName(u.Name))
}

func updateUserFirebase(userId, name, role string, wg *sync.WaitGroup, errchan chan error) {
	defer wg.Done()

	// updating user name
	params := (&auth.UserToUpdate{}).DisplayName(name)
	_, err := firebaseClient.UpdateUser(ctx, userId, params)
	if err != nil {
		log.Fatalf("Error updating user: %v\n", err)
		errchan <- err
		return
	}

	// updating custom claims
	newClaims := map[string]interface{}{"role": role}
	err = firebaseClient.SetCustomUserClaims(ctx, userId, newClaims)
	if err != nil {
		log.Printf("Error setting custom claims: %v\n", err)
	}

	errchan <- err
}

func updateUserDatabase(userId, name, role string, wg *sync.WaitGroup, errchan chan error) {
	defer wg.Done()

	args := dbqueries.UpdateUserArgs(name, role, userId)
	_, err := db.Exec(ctx, dbqueries.UpdateUser, args)
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
	}

	errchan <- err
}

func (u *User) UpdateUser() error {
	errchan := make(chan error, 2)
	wg := new(sync.WaitGroup)

	wg.Add(2)
	go updateUserFirebase(u.UserId, u.Name, u.Role, wg, errchan)
	go updateUserDatabase(u.UserId, u.Name, u.Role, wg, errchan)

	wg.Wait()
	close(errchan)

	for err := range errchan {
		if err != nil {
			return err
		}
	}

	return nil
}

func updateUserNameFirebase(userId, name string, wg *sync.WaitGroup, errchan chan error) {
	defer wg.Done()

	params := (&auth.UserToUpdate{}).DisplayName(name)
	_, err := firebaseClient.UpdateUser(ctx, userId, params)
	if err != nil {
		log.Fatalf("Error updating user: %v\n", err)

	}
	errchan <- err
}

func updateUserNameDatabase(userId, name string, wg *sync.WaitGroup, errchan chan error) {
	defer wg.Done()

	args := dbqueries.UpdateUserNameArgs(name, userId)
	_, err := db.Exec(ctx, dbqueries.UpdateUserName, args)
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
	}

	errchan <- err
}

func (u *User) UpdateUserName() error {
	errchan := make(chan error, 2)
	wg := new(sync.WaitGroup)

	wg.Add(2)
	go updateUserNameFirebase(u.UserId, u.Name, wg, errchan)
	go updateUserNameDatabase(u.UserId, u.Name, wg, errchan)

	wg.Wait()
	close(errchan)

	for err := range errchan {
		if err != nil {
			return err
		}
	}

	return nil
}

/*
delete user
*/
func deleteUserFromFirebase(userId string, wg *sync.WaitGroup, errchan chan error) {
	defer wg.Done()

	err := firebaseClient.DeleteUser(ctx, userId)
	if err != nil {
		log.Printf("Error deleting user from firebase: %v\n", err)
	}
	errchan <- err
}

func deleteUserFromDatabase(userId string, wg *sync.WaitGroup, errchan chan error) {
	// delete user from users table
	// delete user to project mapping for that user
	defer wg.Done()

	args := dbqueries.DeleteUserArgs(userId)
	_, err := db.Exec(ctx, dbqueries.DeleteUser, args)
	if err != nil {
		log.Printf("Error deleting user in database: %v\n", err)
	}
	errchan <- err
}

// delete user
func (u *User) DeleteUser() error {
	errchan := make(chan error, 2)
	wg := new(sync.WaitGroup)

	wg.Add(2)
	go deleteUserFromFirebase(u.UserId, wg, errchan)
	go deleteUserFromDatabase(u.UserId, wg, errchan)

	wg.Wait()
	close(errchan)

	for err := range errchan {
		if err != nil {
			return err
		}
	}

	return nil
}

/*
get user
*/
func (u *User) GetUserById() (*User, error) {
	args := dbqueries.GetUserByIDArgs(u.UserId)
	rows, err := db.Query(ctx, dbqueries.GetUserByID, args)
	if err != nil {
		log.Printf("Error fetching user from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "User with the provided ID does not exist."
			return nil, &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
		}
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &user, nil
}

func (u *User) GetAllUsers() (*[]User, error) {
	rows, err := db.Query(ctx, dbqueries.GetUsers)
	if err != nil {
		log.Printf("Error fetching user from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &users, nil
}
