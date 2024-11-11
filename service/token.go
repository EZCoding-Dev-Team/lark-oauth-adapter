package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AccessTokenService struct {
	larkAppAccessTokenURL string
}

func NewAccessTokenService() *AccessTokenService {
	return &AccessTokenService{
		larkAppAccessTokenURL: "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal",
	}
}

func (s *AccessTokenService) GetAppToken(appId, appSecret string) (string, error) {
	req := struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}{
		AppID:     appId,
		AppSecret: appSecret,
	}

	marshalledReq, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	rawResp, err := http.Post(
		s.larkAppAccessTokenURL,
		"application/json; charset=utf-8",
		bytes.NewReader(marshalledReq),
	)
	if err != nil {
		return "", err
	}

	var resp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		AppAccessToken    string `json:"app_access_token"`
		Expire            int    `json:"expire"`
		TenantAccessToken string `json:"tenant_access_token"`
	}

	err = json.NewDecoder(rawResp.Body).Decode(&resp)
	if err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("failed to get app access token: %s", resp.Msg)
	}

	return resp.AppAccessToken, err
}
