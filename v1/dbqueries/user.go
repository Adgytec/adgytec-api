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

// update user name
const UpdateUserName = `UPDATE users SET name=@name WHERE user_id=@userId`

func UpdateUserNameArgs(name, userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"name":   name,
		"userId": userId,
	}
}

// update user name and role
const UpdateUser = `UPDATE users SET name=@name, role=@role WHERE user_id=@userId`

func UpdateUserArgs(name, role, userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"name":   name,
		"role":   role,
		"userId": userId,
	}
}
