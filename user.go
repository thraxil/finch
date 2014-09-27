package main

import "code.google.com/p/go.crypto/bcrypt"

type User struct {
	Id       int
	Username string
	Password []byte
}

// SetPassword takes a plaintext password and hashes it with bcrypt
func (u *User) SetPassword(password string) string {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) //this is a panic because bcrypt errors on invalid costs
	}
	return string(hpass)
}

func (u User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		return false
	}
	return true
}
