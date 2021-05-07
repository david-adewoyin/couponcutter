package authetication

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
)

var (
	//ErrInvalidEmail is returned if the user provided email address is invalid
	ErrInvalidEmail = errors.New("email is invalid")
	//ErrPasswordLengthUnAcceptable is returned if the user supplied a password length that < 6 || > 64
	ErrPasswordLengthUnAcceptable = errors.New("the lenght of the password is unacceptable")
	//ErrIdentityDoesNotExists is returned if an user cannot ber verified to exists
	ErrIdentityDoesNotExists = errors.New("identity does not exists")
	//ErrIdentityAlreadyExists is returned if an indentity cannot be created because it already exists
	ErrIdentityAlreadyExists = errors.New("identity already exists")
	//ErrUnableToProcessRequest is returned if an authetication request fails
	ErrUnableToProcessRequest = errors.New("unable to process request")
)
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

//User represent an entity to be authenticated
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type service struct {
	repo      Repository
	secretKey string
}

//TokenResponse is returned carrying the token
type TokenResponse struct {
	Token string `json:"token,omitempty"`
}

// NewService returns an Authentication Service Provider
func NewService(repo Repository, secretKey string) Service {
	return &service{repo: repo, secretKey: secretKey}
}

//Repository provides access to storage facilities for authentication
type Repository interface {
	CreateUser(ctx context.Context, email string, password string) (string, error)
	UserWithEmail(ctx context.Context, email string) (string, error)
	UserWithIdentity(ctx context.Context, email string, password string) (string, error)
	ResetPassword(ctx context.Context, email string) error
}

//Service defines the constract for accessing authentication services
type Service interface {
	VerifyToken(authToken string) (string, bool)
	CreateUser(ctx context.Context, email string, password string) (bool, error)
	Login(ctx context.Context, email string, password string) (*TokenResponse, error)
	ResetPassword(ctx context.Context, email string) error
}

//helper function for emails
func trimAndLower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
func (s *service) ResetPassword(ctx context.Context, email string) error {

	valid := isEmailValid(email)
	if !valid {
		return errors.New("password length is not accepted")
	}
	return s.repo.ResetPassword(ctx, email)
}

func (s *service) Login(ctx context.Context, email string, password string) (*TokenResponse, error) {
	email = trimAndLower(email)

	//check if email is of valid type
	valid := isEmailValid(email)
	if !valid {
		return nil, ErrInvalidEmail
	}
	//check if password length if acceptable
	valid = isPasswordLengthValid(password)
	if !valid {
		return nil, ErrPasswordLengthUnAcceptable
	}
	//fetch user with the given email if exists else return error
	userid, err := s.repo.UserWithIdentity(ctx, email, password)
	if err != nil {
		return nil, ErrIdentityDoesNotExists
	}

	token, err := s.createToken(userid)
	if err != nil {
		return nil, ErrUnableToProcessRequest
	}
	return &TokenResponse{Token: token}, nil

}

//TODO  fix bug
func (s *service) CreateUser(ctx context.Context, email string, password string) (bool, error) {
	email = trimAndLower(email)

	//check if email is of valid type
	valid := isEmailValid(email)
	if !valid {
		return false, ErrInvalidEmail
	}
	//check if password length if acceptable
	valid = isPasswordLengthValid(password)
	if !valid {
		return false, ErrPasswordLengthUnAcceptable
	}
	//fetch user with the given email if exists else return error
	_, err := s.repo.UserWithEmail(ctx, email)
	if err == nil {
		return false, ErrIdentityAlreadyExists
	}
	s.repo.CreateUser(ctx, email, password)
	return true, nil

}

// check if the password if of proper length
func isPasswordLengthValid(password string) bool {
	if len(password) < 6 || len(password) > 64 {
		return false
	}
	return true
}

//check if the email address is a valid one
func isEmailValid(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

//create a token from the userid
func (s *service) createToken(userID string) (string, error) {
	claim := jwt.StandardClaims{
		Subject:   userID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 25).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := t.SignedString([]byte(s.secretKey))
	if err != nil {

		return "", errors.New("unable to create token")
	}
	return token, err
}

//create a token from the userid
func (s *service) CreateSignupToken(email string) (string, error) {
	claim := jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := t.SignedString([]byte(s.secretKey))
	if err != nil {

		return "", errors.New("unable to create token")
	}
	return token, err
}

// verify a token
func (s *service) verifytoken(authtoken string) (*jwt.Token, error) {

	token, err := jwt.Parse(authtoken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return "", errors.New("token not properly signed")
		}
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return nil, errors.New("invalid token, signature is false")
	}
	if token.Claims.Valid() != nil {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func (s *service) VerifyToken(authToken string) (string, bool) {

	token, err := s.verifytoken(authToken)
	if err != nil {
		return "", false
	}
	c := token.Claims.(jwt.MapClaims)
	return c["sub"].(string), true

}
