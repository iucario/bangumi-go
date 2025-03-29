package api

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Client *AuthClient
	UserInfo
}

func NewUser(client *AuthClient) *User {
	return &User{
		Client: client,
	}
}

func (u *User) GetUserInfo() (*UserInfo, error) {
	API := "https://api.bgm.tv/v0/me"

	b, err := u.Client.Get(API)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{}
	err = json.Unmarshal(b, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

type UserInfo struct {
	Avatar struct {
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
	} `json:"avatar"`
	Sign      string `json:"sign"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Id        int    `json:"id"`
	UserGroup int    `json:"user_group"`
}

func (u UserInfo) String() string {
	return fmt.Sprintf("User: %s, Nickname: %s, ID: %d", u.Username, u.Nickname, u.Id)
}

type DefaultResponse struct {
	Request string `json:"request"`
	Code    string `json:"code"`
	Error   string `json:"error"`
}
