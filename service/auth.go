package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"lark-oauth-adapter/dto"
	"net/http"
	"net/url"
)

type AuthService struct {
	larkUserAccessTokenURL string
	larkUserInfoURL        string
	larkUserContactInfoURL string
	accessTokenService     *AccessTokenService
}

func NewAuthService() *AuthService {
	return &AuthService{
		larkUserAccessTokenURL: "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token",
		larkUserInfoURL:        "https://open.feishu.cn/open-apis/authen/v1/user_info",
		larkUserContactInfoURL: "https://open.feishu.cn/open-apis/contact/v3/users/%s",
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

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

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
		return nil, fmt.Errorf("failed to get user access token: %d", result.Code)
	}

	return &dto.AccessTokenResponse{
		AccessToken:      result.Data.AccessToken,
		TokenType:        result.Data.TokenType,
		ExpiresIn:        result.Data.ExpiresIn,
		RefreshToken:     result.Data.RefreshToken,
		RefreshExpiresIn: result.Data.RefreshExpiresIn,
	}, nil
}

func (s *AuthService) GetUserInfo(authorization string) (*dto.UserInfo, error) {
	req, err := http.NewRequest("GET", s.larkUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			OpenID          string `json:"open_id"`
			UnionID         string `json:"union_id"`
			Name            string `json:"name"`
			EnterpriseEmail string `json:"enterprise_email"`
			UserID          string `json:"user_id"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get user info: %s", result.Msg)
	}

	data, err := s.getUserDepartmentInfo(authorization, result.Data.UnionID)
	if err != nil {
		return nil, err
	}

	return &dto.UserInfo{
		Sub:               result.Data.UnionID,
		PreferredUsername: result.Data.UserID,
		Name:              result.Data.Name,
		Email:             result.Data.EnterpriseEmail,
		Groups:            data,
	}, nil
}

func (s *AuthService) getUserDepartmentInfo(authorization, unionId string) ([]string, error) {
	uri, err := url.Parse(fmt.Sprintf(s.larkUserContactInfoURL, unionId))
	if err != nil {
		return nil, err
	}
	q := uri.Query()
	q.Set("user_id_type", "union_id")
	q.Set("department_id_type", "department_id")
	uri.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			User struct {
				DepartmentIds []string `json:"department_ids"`
			} `json:"user"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get user detail info: %s", result.Msg)
	}

	return result.Data.User.DepartmentIds, nil
}
