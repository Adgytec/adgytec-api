package dbqueries

import "github.com/jackc/pgx/v5"

// create a single user
const CreateUser = `INSERT INTO users (user_id, name, email, role)
					values (@userId, @name, @email, @role)`

func CreateUserArgs(uid, email, name, role string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId": uid,
		"email":  email,
		"name":   name,
		"role":   role,
	}
}

// get a single user
const GetUserByEmail = `SELECT * FROM users where email=@email`

func GetUserByEmailArgs(email string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"email": email,
	}
}

// delete user by id
const DeleteUser = `DELETE FROM users WHERE user_id=@userId`

func DeleteUserArgs(userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId": userId,
	}
}
