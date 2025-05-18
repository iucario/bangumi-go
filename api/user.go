package api

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Client *AuthClient
	*UserInfo
}

// NewUser creates a new User instance with the provided AuthClient and fetches user info.
func NewUser(client *AuthClient) *User {
	if client == nil {
		return nil
	}
	userInfo, err := getUserInfo(client)
	if err != nil {
		return nil
	}
	return &User{
		Client:   client,
		UserInfo: userInfo,
	}
}

func (u *User) GetUserInfo() (*UserInfo, error) {
	return getUserInfo(u.Client)
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

func getUserInfo(client *AuthClient) (*UserInfo, error) {
	API := "https://api.bgm.tv/v0/me"

	b, err := client.Get(API)
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
