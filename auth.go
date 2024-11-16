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

func (r UserRepostiory) createDB() error {
	_, err := DB.Exec("CREATE TABLE IF NOT EXISTS users (username TEXT, password TEXT);")

	if err != nil {
		return err
	}

	return nil
}

func (r UserRepostiory) add(username string, password string) (int, error) {
	result, err := r.DB.Exec(
		fmt.Sprintf(
			`INSERT INTO users (username, password) VALUES ("%s", "%s");`,
			username, password,
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

func hashPassword(raw_password string) string {
	password, err := bcrypt.GenerateFromPassword([]byte(raw_password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(password)
}

func passwordIsValid(raw_password string, target string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(target), []byte(raw_password))

	if err != nil {
		return false
	}

	return true
}

func (r UserRepostiory) authenticate(username string, raw_password string) (int, error) {
	hashedPassword := hashPassword(raw_password)

	result, err := r.DB.Query(
		fmt.Sprintf(
			`SELECT rowid users WHERE username="%s" AND password="%s"`,
			username, hashedPassword,
		),
	)

	if err != nil {
		return 0, err
	}

	for result.Next() {
		var userid int
		if result.Scan(&userid) != nil {
			return userid, nil
		}
	}

	return 0, errors.New("NO USER FOUND")
}

func (r UserRepostiory) getById(id int) (User, error) {
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
