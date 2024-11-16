package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	RowId    int
	Username string
	Password string
}

type UserRepostiory struct {
	DB *sql.DB
}

func (r UserRepostiory) SetupTable() error {
	InfoLogger.Println("Setting up 'users' table")
	_, err := DB.Exec("CREATE TABLE IF NOT EXISTS users (username TEXT, password TEXT);")

	if err != nil {
		return err
	}

	return nil
}

func (r UserRepostiory) CountAll() int {
	result, err := r.DB.Query("SELECT Count(*) FROM users;")

	if err != nil {
		panic(err)
	}

	var count int

	if !result.Next() {
		panic(errors.New("unable to count users row"))
	}

	err = result.Scan(&count)

	if err != nil {
		panic(err)
	}

	return count

}

func (r UserRepostiory) add(username string, password string) (int, error) {
	hashed_password := hashPassword(password)

	fmt.Println("Add: ", hashed_password)
	result, err := r.DB.Exec(
		fmt.Sprintf(
			`INSERT INTO users (username, password) VALUES ("%s", "%s");`,
			username, hashPassword(password),
		),
	)

	if err != nil {
		return 0, err
	}

	insertedId, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(insertedId), nil
}

func (r UserRepostiory) DeleteById(rowid int) error {
	_, err := r.DB.Exec(
		fmt.Sprintf(
			`DELETE FROM users WHERE rowid=%d;`, rowid,
		),
	)
	return err
}

func hashPassword(raw_password string) string {
	password, err := bcrypt.GenerateFromPassword([]byte(raw_password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(password)
}

func PasswordIsValid(raw_password string, target string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(target), []byte(raw_password))
	return err == nil

}

func (r UserRepostiory) Authenticate(username string, rawPassword string) (int, error) {

	result, err := r.DB.Query(
		fmt.Sprintf(
			`SELECT rowid FROM users WHERE username="%s"`, username,
		),
	)

	if err != nil {
		return 0, err
	}

	var userid int
	for result.Next() {
		if result.Scan(&userid) != nil {
			user, err := r.GetById(userid)

			if err != nil {
				panic(err)
			}

			if !PasswordIsValid(rawPassword, user.Password) {
				return 0, errors.New("invalid crendentials")
			}

			return userid, nil
		}
	}

	return 0, errors.New("no user found with username")
}

func (r *UserRepostiory) GetById(id int) (User, error) {
	result, err := r.DB.Query(
		fmt.Sprintf(
			`SELECT (username, password) users WHERE rowid="%d"`, id,
		),
	)

	if err != nil {
		return User{}, err
	}

	for result.Next() {
		var username string
		var password string
		if result.Scan(&username, &password) != nil {
			user := User{
				Username: username,
				Password: password,
			}

			return user, nil
		}
		break
	}

	return User{}, errors.New("NO USER FOUND")
}

type SessionManager struct {
	DB *sql.DB
}

// Session
func (s *SessionManager) SetupSessionTable() error {
	InfoLogger.Println("Setup Session Table")
	_, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS sessions (user_id INTEGER, key TEXT);`)
	return err
}

func (s *SessionManager) CreateSession(user_id int) (string, error) {
	uuid := uuid.New()
	token := uuid.String()

	query := `INSERT INTO sessions (user_id, key) VALUES ("%d", "%s")`
	_, err := s.DB.Exec(query, user_id, token)

	if err != nil {
		return "", err
	}

	return token, nil
}

var ErrUserNotFound = errors.New("User Not Found")

func (s *SessionManager) FindUserIdByKey(key string) (int, error) {
	query := `SELECT rowid FROM sessions WHERE key="%s"`
	rows, err := s.DB.Query(query, key)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var userId int
		rows.Scan(&userId)

		return userId, nil
	}

	return 0, ErrUserNotFound
}
