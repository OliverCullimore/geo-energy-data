package models

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"log"
)

type Config struct {
	GeoSystemID      string
	ShoutrrrServices []string
}

type Env struct {
	Config                  Config
	DB                      influxdb2.Client
	DBOrg                   string
	DBBucket                string
	Logger                  *log.Logger
	NotificationTemplateDir string
	TemplateDir             string
	StaticDir               string
	ErrorsDir               string
}

/*
type User struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Password   string         `json:"-"`
	Enabled    string         `json:"enabled"`
	LastActive mysql.NullTime `json:"lastactive"`
	Updated    time.Time      `json:"updated"`
	Created    time.Time      `json:"created"`
}

// Create user
func (u *User) Create(env *Env) error {
	// Check if a user already exists with the email
	user := User{Email: u.Email}
	res, err := user.GetByEmail(env)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return err
	}
	if res != nil {
		return errors.New("email already exists")
	}

	// Prepare statement for inserting data
	stmt, err := env.DB.Prepare("INSERT INTO users SET name=?, email=?, password=?, enabled=?")
	// Defer the close and handle any errors
	defer func() {
		cerr := stmt.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.Name, u.Email, u.Password, u.Enabled)
	if err != nil {
		return err
	}

	return nil
}

// Get user by ID
func (u *User) GetByID(env *Env) (*User, error) {
	user := User{}

	err := env.DB.QueryRow("SELECT userid, name, email, password, enabled, lastactive, dateupdated, dateadded FROM users WHERE userid = ?", u.ID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Enabled, &user.LastActive, &user.Updated, &user.Created)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Get user by Email
func (u *User) GetByEmail(env *Env) (*User, error) {
	user := User{}

	err := env.DB.QueryRow("SELECT userid, name, email, password, enabled, lastactive, dateupdated, dateadded FROM users WHERE email = ?", u.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Enabled, &user.LastActive, &user.Updated, &user.Created)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update user
func (u *User) Update(env *Env) error {
	// Check that user exists
	user := User{ID: u.ID}
	res, err := user.GetByID(env)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("user not found")
	}

	// Prepare statement for inserting data
	stmt, err := env.DB.Prepare("UPDATE users SET name=?, email=?, password=?, enabled=?, lastactive=? WHERE userid = ?")
	// Defer the close and handle any errors
	defer func() {
		cerr := stmt.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.Name, u.Email, u.Password, u.Enabled, u.LastActive, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// Delete user
func (u *User) Delete(env *Env) error {
	// Check that user exists
	user := User{ID: u.ID}
	res, err := user.GetByID(env)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("user not found")
	}

	// Prepare statement for deleting data
	stmt, err := env.DB.Prepare("DELETE FROM users WHERE userid = ?")
	// Defer the close and handle any errors
	defer func() {
		cerr := stmt.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetAllUsers retrieves a list of all users
func GetAllUsers(env *Env, start, count int) ([]*User, error) {
	var users []*User

	rows, err := env.DB.Query("SELECT userid, name, email, enabled, lastactive, dateupdated, dateadded FROM users LIMIT ?, ?", start, count)
	// Defer the close and handle any errors
	defer func() {
		cerr := rows.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		u := new(User)
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Enabled, &u.LastActive, &u.Updated, &u.Created)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
*/
