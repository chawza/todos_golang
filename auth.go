package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

	defer result.Close()

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

func (r UserRepostiory) Add(username string, password string) (int, error) {
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

func (r UserRepostiory) CreateSuperUser(username string, password string) (int, error) {
	return r.Add(username, password)
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

	defer result.Close()

	if !result.Next() {
		return 0, errors.New("no user found with username")
	}

	var userid int
	result.Scan(&userid)

	user, err := r.GetById(userid)

	if err != nil {
		panic(err)
	}

	if !PasswordIsValid(rawPassword, user.Password) {
		return 0, errors.New("invalid crendentials")
	}

	return userid, nil
}

func (r *UserRepostiory) GetById(id int) (User, error) {
	result, err := r.DB.Query(
		fmt.Sprintf(
			`SELECT username, password FROM users WHERE rowid=%d`, id,
		),
	)

	if err != nil {
		panic(err)
	}

	defer result.Close()

	var username string
	var password string
	for result.Next() {
		result.Scan(&username, &password)
		user := User{
			Username: username,
			Password: password,
		}
		return user, nil
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
	s.ClearSession(user_id)

	uuid := uuid.New()
	token := uuid.String()

	query := `INSERT INTO sessions (user_id, key) VALUES ($1, "$2");`
	_, err := s.DB.Exec(query, user_id, token)

	if err != nil {
		return "", err
	}

	return token, nil
}
func (s *SessionManager) ClearSession(user_id int) {
	query := `SELECT rowid FROM sessions WHERE user_id=$1;`
	result, err := s.DB.Query(query, user_id)

	if err != nil {
		panic(err)
	}

	result.Close()

	isCached := result.Next()

	if !isCached {
		return
	}

	_, err = s.DB.Exec(`DELETE FROM sessions WHERE user_id=$1;`, user_id)

	if err != nil {
		panic(err)
	}
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

func PromptToCreateSuperUser() {
	InfoLogger.Println("Setup Admin")

	var inputPasword string
	fmt.Print("Admin password: ")
	fmt.Scan(&inputPasword)

	adminId, err := UserRepo.Add("admin", inputPasword)

	if err != nil {
		panic(err)
	}

	admin, err := UserRepo.GetById(adminId)

	if err != nil {
		panic(err)
	}

	InfoLogger.Println("Admin created by username: ", admin.Username)
}

type AuthApi struct {
}

// API
func (*AuthApi) Login(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	username := q.Get("username")
	password := q.Get("password")

	if len(username) == 0 || len(password) == 0 {
		WriteClientError(w, ClientError{Message: "Credential is required"})
		return
	}

	user_id, err := UserRepo.Authenticate(username, password)

	if err != nil {
		ErrorLogger.Println(`Auth: `, err.Error())
		WriteClientError(w, ClientError{Message: "Credential is Invalid"})
		return
	}

	user_session, err := UserSession.CreateSession(user_id)

	if err != nil {
		panic(err)
	}

	type LoginData struct {
		Token string `json:"token"`
	}

	bodyPayload, err := json.Marshal(
		struct {
			Data LoginData `json:"data"`
		}{
			Data: LoginData{
				Token: user_session,
			},
		},
	)

	if err != nil {

	}

	w.Write(bodyPayload)
}
