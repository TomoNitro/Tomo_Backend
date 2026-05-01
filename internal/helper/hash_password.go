package helper

import "github.com/go-crypt/x/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
func CompareHashPasswordAndPassword(password, userPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(userPassword))
}
