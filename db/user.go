package db

type User struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
