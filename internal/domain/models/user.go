package models

type User struct {
	Id       int64  `json:"id" db:"id"`
	Email    string `json:"email" db:"user_email"`
	PassHash []byte `json:"password_hash" db:"pass_hash"`
}
