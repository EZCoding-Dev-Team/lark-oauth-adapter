package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lark-oauth-adapter/dto"
	"net/http"
)

type AuthService struct {
	larkUserAccessTokenURL string
	larkUserInfoURL        string
	accessTokenService     *AccessTokenService
}

func NewAuthService() *AuthService {
	return &AuthService{
		larkUserAccessTokenURL: "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token",
		larkUserInfoURL:        "https://open.feishu.cn/open-apis/authen/v1/user_info",
		accessTokenService:     NewAccessTokenService(),
	}
}

func (s *AuthService) GetAccessToken(data dto.AccessTokenRequest) (*dto.AccessTokenResponse, error) {
	authToken, err := s.accessTokenService.GetAppToken(data.ClientID, data.ClientSecret)
	if err != nil {
		return nil, err
	}

	payload := struct {
		Code      string `json:"code"`
		GrantType string `json:"grant_type"`
	}{
		Code:      data.Code,
		GrantType: "authorization_code",
	}
	marshaledData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.larkUserAccessTokenURL, bytes.NewBuffer(marshaledData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AccessToken      string `json:"access_token"`
			RefreshToken     string `json:"refresh_token"`
			TokenType        string `json:"token_type"`
			ExpiresIn        int    `json:"expires_in"`
			RefreshExpiresIn int    `json:"refresh_expires_in"`
			Scope            string `json:"scope"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get user access token: %s", result.Msg)
	}

	return &dto.AccessTokenResponse{
		AccessToken:      result.Data.AccessToken,
		TokenType:        result.Data.TokenType,
		ExpiresIn:        result.Data.ExpiresIn,
		RefreshToken:     result.Data.RefreshToken,
		RefreshExpiresIn: result.Data.RefreshExpiresIn,
	}, nil
}

func (s *AuthService) GetUserInfo(authorization string) (map[string]any, error) {
	req, err := http.NewRequest("GET", s.larkUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data map[string]any `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get user info: %s", result.Msg)
	}

	return result.Data, nil
}
