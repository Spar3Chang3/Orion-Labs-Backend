package serverauth

import (
	"crypto/rand"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var accessMap = make(map[string]*User)
var userMap = make(map[string]*User)

type AccessToken struct {
	Token   string `json:"token"`
	TokenIP string `json:"tokenIP"`
}

type User struct {
	UserName        string        `json:"userName"`
	PasswordHash    string        `json:"passwordHash"`
	AccessTokens    []AccessToken `json:"accessTokens"`
	AccessTokenDate time.Time     `json:"accessTokenDate"`
	LastAccessDate  time.Time     `json:"lastAccessDate"`
}

func init() {
	//implement grabbing all the users from /data/auth
}

func generateNewAccessToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	token := b64.URLEncoding.EncodeToString(tokenBytes)
	return token, nil
}

func getUserByToken(token string) (*User, error) {
	user, exists := accessMap[token]
	if !exists {
		return nil, errors.New("access token not found")
	}
	return user, nil
}

func checkDuplicateUser(UserName string) bool { //ed participates
	_, exists := accessMap[UserName]
	return exists
}

func checkUserLogin(UserName string, encodedPassword string) (bool, error) {
	user, err := userMap[UserName]
	if err {
		return false, errors.New("could not validate user login")
	}

	decodedPassword, _ := b64.StdEncoding.DecodeString(encodedPassword)
	bytePassword := []byte(decodedPassword)

	passErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), bytePassword)

	if passErr != nil {
		return false, errors.New("incorrect username and/or password")
	}

	return true, nil
}

func addNewUser(UserName string, encodedPassword string, ip string) error {
	if checkDuplicateUser(UserName) == true {
		return errors.New("duplicate username detected")
	}
	PasswordHash := encrypt(encodedPassword)
	token, err := generateNewAccessToken()
	if err != nil {
		return err
	}

	userPointer := &User{
		UserName:     UserName,
		PasswordHash: string(PasswordHash),
		AccessTokens: []AccessToken{
			{
				Token:   token,
				TokenIP: ip,
			},
		},
		AccessTokenDate: time.Now(),
		LastAccessDate:  time.Now(),
	}

	userMap[UserName] = userPointer
	accessMap[token] = userPointer

	return nil
}

func encrypt(encodedPassword string) []byte {
	pass, _ := b64.StdEncoding.DecodeString(encodedPassword)

	encryptedPass, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		fmt.Errorf("failed to encrypt password with cost: %d", bcrypt.DefaultCost)
	}
	return encryptedPass
}

//func encryptAndStore(encodedPassword string, userId string) {
//	pass := encrypt(encodedPassword)
//
//}
