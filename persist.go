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

type persistence struct {
	Database *sql.DB
}

func newPersistence(dbfile string) *persistence {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return &persistence{Database: db}
}

func (p *persistence) Close() {
	p.Database.Close()
}

func (p persistence) GetUser(username string) (*user, error) {
	stmt, err := p.Database.Prepare("select id, password from users where username = ?")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var id int
	var password string

	err = stmt.QueryRow(username).Scan(&id, &password)
	if err != nil {
		return nil, err
	}
	return &user{ID: id, Username: username, Password: []byte(password)}, err
}

func (p persistence) getUserByID(id int) (*user, error) {
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
	return &user{ID: id, Username: username, Password: []byte(password)}, nil
}

func (p *persistence) CreateUser(username, password string) (*user, error) {
	var user user
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

type channel struct {
	ID    int
	User  *user
	Slug  string
	Label string
}

func (p persistence) GetUserChannels(u user) ([]*channel, error) {
	q := `select id, slug, label from channel where user_id = ? order by slug asc`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var channels []*channel

	rows, err := stmt.Query(u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var slug string
		var label string
		rows.Scan(&id, &slug, &label)
		c := &channel{ID: id, Slug: slug, Label: label}
		channels = append(channels, c)
	}
	return channels, nil
}

func (p *persistence) AddChannels(u user, names []string) ([]*channel, error) {
	var created []*channel
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
		r, err := stmt.Exec(u.ID, slug, label)

		id, err := r.LastInsertId()
		if err != nil {
			log.Println("error getting last inserted id", err)
			return nil, err
		}
		c := &channel{ID: int(id), Slug: slug, Label: label, User: &u}
		created = append(created, c)
	}

	tx.Commit()
	return created, nil
}

func (p *persistence) DeleteChannel(c *channel) error {
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
	_, err = stmt.Exec(c.ID)
	if err != nil {
		return err
	}

	stmt2, err := tx.Prepare(q2)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer stmt2.Close()
	_, err = stmt2.Exec(c.ID)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (p *persistence) DeletePost(post *post) error {
	q1 := `delete from postchannel where post_id = ?`
	q2 := `delete from post where id = ?`

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
	_, err = stmt.Exec(post.ID)
	if err != nil {
		return err
	}

	stmt2, err := tx.Prepare(q2)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer stmt2.Close()
	_, err = stmt2.Exec(post.ID)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (p persistence) GetChannel(u user, slug string) (*channel, error) {
	q := `select id, label from channel where user_id = ? AND slug = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var id int
	var label string

	err = stmt.QueryRow(u.ID, slug).Scan(&id, &label)
	if err != nil {
		return nil, err
	}
	return &channel{ID: id, User: &u, Slug: slug, Label: label}, nil
}

func (p persistence) GetChannelByID(id int) (*channel, error) {
	q := `select user_id, slug, label from channel where id = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var slug string
	var label string
	var userID int

	err = stmt.QueryRow(id).Scan(&userID, &slug, &label)
	if err != nil {
		return nil, err
	}

	u, err := p.getUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &channel{ID: id, User: u, Slug: slug, Label: label}, nil
}

func (p persistence) getPost(id int) (*post, error) {
	q := `select user_id, uuid, body, posted from post where id = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var body string
	var userID int
	var posted int
	var uu string

	err = stmt.QueryRow(id).Scan(&userID, &uu, &body, &posted)
	if err != nil {
		log.Println("error querying by post id", err)
		return nil, err
	}

	u, err := p.getUserByID(userID)
	if err != nil {
		log.Println("error getting post user", err)
		return nil, err
	}
	// TODO: also get channels
	return &post{ID: id, UUID: uu, User: u, Body: body, Posted: posted}, nil
}

func (p persistence) GetPostByUUID(uu string) (*post, error) {
	q := `select id, user_id, body, posted from post where uuid = ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var body string
	var id int
	var userID int
	var posted int

	err = stmt.QueryRow(uu).Scan(&id, &userID, &body, &posted)
	if err != nil {
		log.Println("error querying by post id", err)
		return nil, err
	}

	u, err := p.getUserByID(userID)
	if err != nil {
		log.Println("error getting post user", err)
		return nil, err
	}
	// TODO: also get channels
	return &post{ID: id, UUID: uu, User: u, Body: body, Posted: posted}, nil
}

func (p persistence) GetAllPosts(limit int, offset int) ([]*post, error) {
	q := `select id, uuid, user_id, body, posted
        from post order by posted desc limit ? offset ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var posts []*post

	rows, err := stmt.Query(limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var userID int
		var body string
		var posted int
		var uu string
		rows.Scan(&id, &uu, &userID, &body, &posted)
		u, err := p.getUserByID(userID)
		if err != nil {
			continue
		}
		post := &post{ID: id, UUID: uu, User: u, Body: body, Posted: posted}
		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels

		posts = append(posts, post)
	}
	return posts, nil
}

func (p persistence) GetPostChannels(post *post) ([]*channel, error) {
	q := `select c.id, c.label, c.slug
        from channel c, postchannel pc
        where pc.channel_id = c.id
          and pc.post_id = ?
        order by c.slug asc`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var channels []*channel

	rows, err := stmt.Query(post.ID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var label string
		var slug string
		rows.Scan(&id, &label, &slug)
		channel := &channel{ID: id, User: post.User, Label: label, Slug: slug}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (p persistence) GetAllPostsInChannel(c channel, limit int, offset int) ([]*post, error) {
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

	var posts []*post

	rows, err := stmt.Query(c.ID, limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var userID int
		var body string
		var posted int
		var uu string
		rows.Scan(&id, &uu, &userID, &body, &posted)
		u, err := p.getUserByID(userID)
		if err != nil {
			continue
		}
		post := &post{ID: id, UUID: uu, User: u, Body: body, Posted: posted}

		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels
		posts = append(posts, post)
	}
	return posts, nil

}

func (p persistence) SearchPosts(query string, limit int, offset int) ([]*post, error) {
	q := `select p.id, p.uuid, p.user_id, p.body, p.posted
        from post p
        where p.body like ?
        order by p.posted desc limit ? offset ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var posts []*post

	rows, err := stmt.Query("%"+query+"%", limit, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var userID int
		var body string
		var posted int
		var uu string
		rows.Scan(&id, &uu, &userID, &body, &posted)
		u, err := p.getUserByID(userID)
		if err != nil {
			continue
		}
		post := &post{ID: id, UUID: uu, User: u, Body: body, Posted: posted}

		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels
		posts = append(posts, post)
	}
	return posts, nil

}

func (p persistence) GetAllUserPosts(u *user, limit int, offset int) ([]*post, error) {
	q := `select id, uuid, body, posted
        from post where user_id = ? order by posted desc limit ? offset ?`
	stmt, err := p.Database.Prepare(q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer stmt.Close()

	var posts []*post

	rows, err := stmt.Query(u.ID, limit, offset)
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
		post := &post{ID: id, UUID: uu, User: u, Body: body, Posted: posted}
		channels, err := p.GetPostChannels(post)

		if err != nil {
			return nil, err
		}
		post.Channels = channels
		posts = append(posts, post)
	}
	return posts, nil
}

func (p *persistence) AddPost(u user, body string, channels []*channel) (*post, error) {
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

	r, err := stmt.Exec(u.ID, u4.String(), body, time.Now().Unix())
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
		_, err = cstmt.Exec(int(id), c.ID)
		if err != nil {
			log.Println("error associating channel with post", err)
		}
	}

	tx.Commit()

	post, err := p.getPost(int(id))
	if err != nil {
		log.Println("error getting post", err)
		return nil, err
	}

	return post, nil
}
