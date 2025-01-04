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

func addNewUserToken(user User, ip string) (AccessToken, error) {
	newToken, err := generateNewAccessToken()
	if err != nil {
		return AccessToken{}, err
	}

	newAccess := AccessToken{
		Token:   newToken,
		TokenIP: ip,
	}

	user.AccessTokens = append(user.AccessTokens, newAccess)

	return newAccess, nil
}

func checkDuplicateUser(UserName string) bool { //ed participates
	_, exists := accessMap[UserName]
	return exists
}

func checkUserTokenLogin(token string, ip string) (bool, error) {
	user, exists := accessMap[token]
	if !exists {
		return false, errors.New("access token not found")
	}
	//We should probably come up with a better way, but right now the map only provides a user pointer from token
	//This means the AccessToken array still needs to be iterated over match said token
	//However, how many tokens will there reasonably be? A max of 10 maybe? eh, sounds like too much refactoring
	tokenArray := user.AccessTokens

	for _, accessToken := range tokenArray { //this means for [however long this obj is], iterate with accessToken as current index
		if accessToken.Token == token && accessToken.TokenIP == ip {
			return true, nil
		}
	}

	return false, errors.New("associated IP no longer valid")
}

func checkUserLogin(UserName string, encodedPassword string, ip string) (bool, error, AccessToken) {
	user, exists := userMap[UserName]
	if !exists {
		return false, errors.New("could not validate user login"), AccessToken{}
	}

	decodedPassword, _ := b64.StdEncoding.DecodeString(encodedPassword)
	bytePassword := []byte(decodedPassword)

	passErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), bytePassword)

	if passErr != nil {
		return false, errors.New("incorrect username and/or password"), AccessToken{}
	}

	newAccess, err := addNewUserToken(*user, ip)

	if err != nil {
		return false, err, AccessToken{}
	}

	return true, nil, newAccess
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
