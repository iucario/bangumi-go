package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

type DefaultResponse struct {
	Request string `json:"request"`
	Code    string `json:"code"`
	Error   string `json:"error"`
}

func (u UserInfo) String() string {
	return fmt.Sprintf("User: %s, Nickname: %s, ID: %d", u.Username, u.Nickname, u.Id)
}

func GetUserInfo(accessToken string) (UserInfo, error) {
	API := "https://api.bgm.tv/v0/me"

	req, err := http.NewRequest("GET", API, nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	if res.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("[error] status code: %d, response: %s", res.StatusCode, bodyString)
	}

	userInfo := UserInfo{}
	err = json.Unmarshal(bodyBytes, &userInfo)
	if err != nil {
		return UserInfo{}, err
	}

	return userInfo, nil
}
