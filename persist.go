package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nu7hatch/gouuid"
)

type Persistence struct {
	Database *sql.DB
}

func NewPersistence(dbfile string) *Persistence {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return &Persistence{Database: db}
}

func (p *Persistence) Close() {
	p.Database.Close()
}

func (p Persistence) GetUser(username string) (*User, bool) {
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

func (p Persistence) GetUserById(id int) (*User, error) {
	q := `select username, password from users where id = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var username string
	var password string

	err = stmt.QueryRow(id).Scan(&username, &password)
	if err != nil {
		return nil, err
	}
	return &User{Id: id, Username: username, Password: []byte(password)}, nil
}

func (p *Persistence) CreateUser(username, password string) (*User, error) {
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

func (p Persistence) GetAllUsers() []*User {
	q := `select id, username from users order by username ASC`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer stmt.Close()

	users := make([]*User, 0)

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var username string
		rows.Scan(&id, &username)
		u := &User{Id: id, Username: username}
		users = append(users, u)
	}
	return users
}

type Channel struct {
	Id    int
	User  *User
	Slug  string
	Label string
}

func (p Persistence) UserChannels(u User) []*Channel {
	q := `select id, slug, label from channel where user_id = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer stmt.Close()

	channels := make([]*Channel, 0)

	rows, err := stmt.Query(u.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var slug string
		var label string
		rows.Scan(&id, &slug, &label)
		c := &Channel{Id: id, Slug: slug, Label: label}
		channels = append(channels, c)
	}
	return channels
}

func (p *Persistence) AddChannels(u User, names []string) ([]*Channel, error) {
	created := make([]*Channel, 0)
	tx, err := p.Database.Begin()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	q := `insert into channel(user_id, slug, label) values(?, ?, ?)`
	stmt, err := tx.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()
	for _, label := range names {
		if label == "" {
			continue
		}
		slug := strings.ToLower(strings.Replace(label, " ", "_", -1))
		r, err := stmt.Exec(u.Id, slug, label)

		id, err := r.LastInsertId()
		if err != nil {
			log.Println("error getting last inserted id", err)
			return nil, err
		}
		c := &Channel{Id: int(id), Slug: slug, Label: label, User: &u}
		created = append(created, c)
	}

	tx.Commit()
	return created, nil
}

func (p *Persistence) DeleteChannel(c *Channel) error {
	q1 := `delete from postchannel where channel_id = ?`
	q2 := `delete from channel where id = ?`

	tx, err := p.Database.Begin()
	if err != nil {
		log.Fatal(err)
		return err
	}
	stmt, err := tx.Prepare(q1)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(c.Id)
	if err != nil {
		return err
	}

	stmt2, err := tx.Prepare(q2)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer stmt2.Close()
	_, err = stmt2.Exec(c.Id)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (p Persistence) GetChannel(u User, slug string) (*Channel, error) {
	q := `select id, label from channel where user_id = ? AND slug = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var id int
	var label string

	err = stmt.QueryRow(u.Id, slug).Scan(&id, &label)
	if err != nil {
		return nil, err
	}
	return &Channel{Id: id, User: &u, Slug: slug, Label: label}, nil
}

func (p Persistence) GetChannelById(id int) (*Channel, error) {
	q := `select user_id, slug, label from channel where id = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var slug string
	var label string
	var user_id int

	err = stmt.QueryRow(id).Scan(&user_id, &slug, &label)
	if err != nil {
		return nil, err
	}

	u, err := p.GetUserById(user_id)
	if err != nil {
		return nil, err
	}

	return &Channel{Id: id, User: u, Slug: slug, Label: label}, nil
}

func (p Persistence) GetPost(id int) (*Post, error) {
	q := `select user_id, uuid, body, posted from post where id = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var body string
	var user_id int
	var posted int
	var uu string

	err = stmt.QueryRow(id).Scan(&user_id, &uu, &body, &posted)
	if err != nil {
		log.Println("error querying by post id", err)
		return nil, err
	}

	u, err := p.GetUserById(user_id)
	if err != nil {
		log.Println("error getting post user", err)
		return nil, err
	}
	// TODO: also get channels
	return &Post{Id: id, UUID: uu, User: u, Body: body, Posted: posted}, nil
}

func (p Persistence) GetPostByUUID(uu string) (*Post, error) {
	q := `select id, user_id, body, posted from post where uuid = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var body string
	var id int
	var user_id int
	var posted int

	err = stmt.QueryRow(uu).Scan(&id, &user_id, &body, &posted)
	if err != nil {
		log.Println("error querying by post id", err)
		return nil, err
	}

	u, err := p.GetUserById(user_id)
	if err != nil {
		log.Println("error getting post user", err)
		return nil, err
	}
	// TODO: also get channels
	return &Post{Id: id, UUID: uu, User: u, Body: body, Posted: posted}, nil
}

func (p Persistence) GetAllPosts(limit int, offset int) ([]*Post, error) {
	q := `select id, uuid, user_id, body, posted
        from post order by posted desc limit ? offset ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	posts := make([]*Post, 0)

	rows, err := stmt.Query(limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var user_id int
		var body string
		var posted int
		var uu string
		rows.Scan(&id, &uu, &user_id, &body, &posted)
		u, err := p.GetUserById(user_id)
		if err != nil {
			continue
		}
		post := &Post{Id: id, UUID: uu, User: u, Body: body, Posted: posted}
		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels

		posts = append(posts, post)
	}
	return posts, nil
}

func (p Persistence) GetPostChannels(post *Post) ([]*Channel, error) {
	q := `select c.id, c.label, c.slug
        from channel c, postchannel pc
        where pc.channel_id = c.id
          and pc.post_id = ?
        order by c.label asc`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	channels := make([]*Channel, 0)

	rows, err := stmt.Query(post.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var label string
		var slug string
		rows.Scan(&id, &label, &slug)
		channel := &Channel{Id: id, User: post.User, Label: label, Slug: slug}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (p Persistence) GetAllPostsInChannel(c Channel, limit int, offset int) ([]*Post, error) {
	q := `select p.id, p.uuid, p.user_id, p.body, p.posted
        from post p, postchannel pc
        where p.id = pc.post_id
          and pc.channel_id = ?
        order by p.posted desc limit ? offset ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	posts := make([]*Post, 0)

	rows, err := stmt.Query(c.Id, limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var user_id int
		var body string
		var posted int
		var uu string
		rows.Scan(&id, &uu, &user_id, &body, &posted)
		u, err := p.GetUserById(user_id)
		if err != nil {
			continue
		}
		post := &Post{Id: id, UUID: uu, User: u, Body: body, Posted: posted}

		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels
		posts = append(posts, post)
	}
	return posts, nil

}

func (p Persistence) GetAllUserPosts(u *User, limit int, offset int) ([]*Post, error) {
	q := `select id, uuid, body, posted
        from post where user_id = ? order by posted desc limit ? offset ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	posts := make([]*Post, 0)

	rows, err := stmt.Query(u.Id, limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var body string
		var posted int
		var uu string
		rows.Scan(&id, &uu, &body, &posted)
		post := &Post{Id: id, UUID: uu, User: u, Body: body, Posted: posted}
		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels
		posts = append(posts, post)
	}
	return posts, nil
}

func (p *Persistence) AddPost(u User, body string, channels []*Channel) (*Post, error) {
	tx, err := p.Database.Begin()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	u4, err := uuid.NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}

	q := `insert into post(user_id, uuid, body, posted) values(?, ?, ?, ?)`
	stmt, err := tx.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	r, err := stmt.Exec(u.Id, u4.String(), body, time.Now().Unix())
	if err != nil {
		log.Println("error inserting post", err)
		return nil, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		log.Println("error getting last inserted id", err)
		return nil, err
	}

	q2 := `insert into postchannel (post_id, channel_id) values (?, ?)`
	cstmt, err := tx.Prepare(q2)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer cstmt.Close()
	for _, c := range channels {
		if c == nil {
			continue
		}
		_, err = cstmt.Exec(int(id), c.Id)
		if err != nil {
			log.Println("error associating channel with post", err)
		}
	}

	tx.Commit()

	post, err := p.GetPost(int(id))
	if err != nil {
		log.Println("error getting post", err)
		return nil, err
	}

	return post, nil
}
