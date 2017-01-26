package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type User struct {
	Email    string
	Hash     string
	Username string
}

var users []User

const userFile = "users.json"

func GetUserHash(email, password string) (string, error) {
	err := readUsers()
	if err != nil {
		return "", err
	}
	for _, u := range users {
		if u.Email == email && u.Hash == fmt.Sprintf("%x", sha256.Sum256([]byte(email+password))) {
			return u.Hash, nil
		}
	}
	return "", errors.New("the user is not exists")
}

func GetUserName(hash string) (string, error) {
	err := readUsers()
	if err != nil {
		return "", err
	}
	for _, u := range users {
		if u.Hash == hash {
			return u.Username, nil
		}
	}
	return "", errors.New("the user is not exists")
}

func readUsers() error {
	if len(users) == 0 {
		_, err := os.Stat(userFile)
		if err != nil {
			f, _ := os.Create(userFile)
			f.Close()
		}
		f, err := os.Open(userFile)
		defer f.Close()
		if err != nil {
			return err
		}
		r := bufio.NewReader(f)
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		json.Unmarshal(b, &users)
	}
	return nil
}

func RegisterUser(user, email, password string) error {
	err := readUsers()
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Email == email {
			fmt.Println(fmt.Sprintf("%x", sha256.Sum256([]byte(email+password))))
			return errors.New("the user already exists")
		}
	}
	new := User{
		Email:    email,
		Username: user,
		Hash:     fmt.Sprintf("%x", sha256.Sum256([]byte(email+password))),
	}
	users = append(users, new)
	b, err := json.Marshal(users)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("users.json", os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		fmt.Println("failed")
		return err
	}
	return nil
}
