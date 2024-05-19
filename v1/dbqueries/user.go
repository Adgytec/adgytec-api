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

// get all users
const GetUsers = `Select * FROM users where cursor>@cursor LIMIT 100`

func GetUsersArgs(cursor int) pgx.NamedArgs {
	return pgx.NamedArgs{
		"cursor": cursor,
	}
}

// get a single user by email
const GetUserByEmail = `SELECT * FROM users where email=@email`

func GetUserByEmailArgs(email string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"email": email,
	}
}

//get a single user by userid
const GetUserByID = `SELECT * FROM users where user_id=@userId`

func GetUserByIDArgs(userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId": userId,
	}
}

// delete user by id
// const DeleteUser = `DELETE FROM users WHERE user_id=@userId`
const DeleteUser = `WITH deleted_user AS(
	DELETE FROM users 
	WHERE user_id=@userId
	RETURNING user_id
)
DELETE FROM user_to_project
WHERE user_id IN (SELECT user_id FROM deleted_user)`

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
