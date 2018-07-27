package main

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type UserRegister struct {
	Username  string `validate:"required"`
	Email     string `validate:"required,email"`
	Password  string `validate:"required"`
	CreatedAt time.Time
}

type UserLogin struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type UserDetails struct {
	ID       string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type User struct {
	ID          string `validate:"required"`
	Email       string `validate:"required,email"`
	Password    string `validate:"required"`
	AccessToken string
	CreatedAt   string
}

var validate *validator.Validate

func Register(w http.ResponseWriter, r *http.Request) {

	var user UserRegister
	_ = json.NewDecoder(r.Body).Decode(&user)

	// Validate

	validate = validator.New()

	err := validate.Struct(user)

	if err != nil {

		n := len(err.(validator.ValidationErrors))
		validateArray := make([]string, n)
		for index, err := range err.(validator.ValidationErrors) {
			validateArray[index] = err.Field() + ":Field validation for " + err.Field() + " failed on the " + err.Tag() + " tag"
		}

		mapArr := map[string]interface{}{"validationError": validateArray}
		resError, _ := json.Marshal(mapArr)

		// Return success response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resError)
		return
	}

	// Open Databse

	db := OpenDB()

	stmt, err := db.Prepare("INSERT INTO users(id, email, password, created_at) VALUES(?,?,?, NOW())")

	// Response Error

	if err != nil {

		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500, "message": ` + err.Error() + `}`)
		w.Write(resErr)
		return
	}

	// Create UUID
	UUID := uuid.NewV4()

	// Hash Password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)

	// Store to database
	stmt.Exec(UUID, user.Email, hashPassword)
	defer db.Close()

	// Return success response
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	resSuccess := []byte(`{"message": "Success"}`)
	w.Write(resSuccess)

}

func Login(w http.ResponseWriter, r *http.Request) {

	var user UserLogin
	_ = json.NewDecoder(r.Body).Decode(&user)

	// Validate
	validate = validator.New()

	err := validate.Struct(user)

	if err != nil {

		n := len(err.(validator.ValidationErrors))
		validateArray := make([]string, n)
		for index, err := range err.(validator.ValidationErrors) {
			validateArray[index] = err.Field() + ":Field validation for " + err.Field() + " failed on the " + err.Tag() + " tag"
		}

		mapArr := map[string]interface{}{"code": 422, "validationError": validateArray}
		resError, _ := json.Marshal(mapArr)

		// Return success response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resError)
		return
	}

	// Get User Details
	var checkUser UserDetails

	db := OpenDB()

	err = db.QueryRow("SELECT id, email,password FROM users WHERE email = ?", user.Email).Scan(&checkUser.ID, &checkUser.Email, &checkUser.Password)
	if err != nil {
		// Return Unauthorized response
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 401,  message": "Unauthorized"}`)
		w.Write(resErr)
		return
	}

	matchPass := CheckPasswordHash(user.Password, checkUser.Password)
	if matchPass {

		// Create Access Token

		tokenInt := string(rand.Int())
		hashIntToken, _ := bcrypt.GenerateFromPassword([]byte(tokenInt), 10)
		accessTokenToB64 := base64.StdEncoding.EncodeToString([]byte(hashIntToken))

		// Store to database
		db := OpenDB()

		stmt, err := db.Prepare("INSERT INTO access_tokens(id, user_id, access_token, created_at) VALUES(?,?,?, NOW())")
		if err != nil {
			// Return success response
			w.WriteHeader(401)
			w.Header().Set("Content-Type", "application/json")
			resErr := []byte(`{"code": 401,  message": "` + err.Error() + `"}`)
			w.Write(resErr)
			return
		}

		stmt.Exec("1", checkUser.ID, hashIntToken)
		defer db.Close()

		// Return success response
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		resSuccess := []byte(`{"access_token": ` + accessTokenToB64 + ` }`)
		w.Write(resSuccess)
		return

	} else {
		// Return success response
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 401, "message":  "Unauthorized"}`)
		w.Write(resErr)
		return
	}

}

func CheckAuthenticated(accessToken string) bool {

	var accessTok string

	// Decode BS64
	accessTokenToB64, _ := base64.StdEncoding.DecodeString(accessToken)
	toString := string(accessTokenToB64)
	db := OpenDB()
	err := db.QueryRow("SELECT access_token FROM access_tokens WHERE access_token = ?", toString).Scan(&accessTok)
	return err == nil

}

func Auth(accessToken, cloumn string) string {
	var user User

	accessTokenToB64, _ := base64.StdEncoding.DecodeString(accessToken)
	toString := string(accessTokenToB64)
	db := OpenDB()
	db.QueryRow("SELECT users.id, users.email, users.password, users.created_at, access_tokens.access_token FROM `access_tokens` JOIN users ON access_tokens.user_id = users.id  WHERE access_token = ?", toString).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.AccessToken)

	if cloumn == "id" {
		return user.ID

	} else if cloumn == "email" {
		return user.Email

	} else if cloumn == "password" {
		return user.Password

	} else if cloumn == "created_at" {
		return user.CreatedAt

	} else if cloumn == "access_token" {
		return user.AccessToken
	}

	return ""

}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
