package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Persistance struct {
	Database *sql.DB
}

func NewPersistance(dbfile string) *Persistance {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return &Persistance{Database: db}
}

func (p *Persistance) Close() {
	p.Database.Close()
}

func (p Persistance) GetUser(username string) (*User, bool) {
	stmt, err := p.Database.Prepare("select id, password from users where username = ?")
	if err != nil {
		log.Fatal(err)
		return nil, false
	}
	defer stmt.Close()

	var id int
	var password string

	err = stmt.QueryRow(username).Scan(&id, &password)
	if err != nil {
		return nil, false
	}
	return &User{Id: id, Username: username, Password: []byte(password)}, true
}

func (p *Persistance) CreateUser(username, password string) (*User, error) {
	var user User
	user.Username = username
	encpassword := user.SetPassword(password)

	tx, err := p.Database.Begin()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	stmt, err := tx.Prepare("insert into users(username, password) values(?, ?)")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, encpassword)
	tx.Commit()

	u, _ := p.GetUser(username)
	return u, nil
}
