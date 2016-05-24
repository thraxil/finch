package main

import "golang.org/x/crypto/bcrypt"

type user struct {
	ID       int
	Username string
	Password []byte
}

// SetPassword takes a plaintext password and hashes it with bcrypt
func (u *user) SetPassword(password string) string {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) //this is a panic because bcrypt errors on invalid costs
	}
	return string(hpass)
}

func (u user) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		return false
	}
	return true
}
