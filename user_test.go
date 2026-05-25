package main

import (
	"testing"
)

func TestUserPassword(t *testing.T) {
	u := &user{
		ID:       1,
		Username: "testuser",
	}

	password := "supersecret"
	hashedPassword := u.SetPassword(password)

	if hashedPassword == "" {
		t.Error("SetPassword returned empty string")
	}

	u.Password = []byte(hashedPassword)

	if !u.CheckPassword(password) {
		t.Error("CheckPassword failed for correct password")
	}

	if u.CheckPassword("wrongpassword") {
		t.Error("CheckPassword succeeded for incorrect password")
	}
}
