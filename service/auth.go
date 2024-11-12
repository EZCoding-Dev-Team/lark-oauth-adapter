package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"io"
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

	log.Debugf("User access token: %s", result.Data.AccessToken)

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
			Name            string `json:"name"`
			EnName          string `json:"en_name"`
			AvatarURL       string `json:"avatar_url"`
			AvatarThumb     string `json:"avatar_thumb"`
			AvatarMiddle    string `json:"avatar_middle"`
			AvatarBig       string `json:"avatar_big"`
			OpenID          string `json:"open_id"`
			UnionID         string `json:"union_id"`
			Email           string `json:"email"`
			EnterpriseEmail string `json:"enterprise_email"`
			UserID          string `json:"user_id"`
			Mobile          string `json:"mobile"`
			TenantKey       string `json:"tenant_key"`
			EmployeeNo      string `json:"employee_no"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get user info: %s", result.Msg)
	}

	return &dto.UserInfo{
		Sub:               result.Data.UnionID,
		PreferredUsername: result.Data.UserID,
		Name:              result.Data.Name,
		EnName:            result.Data.EnName,
		AvatarURL:         result.Data.AvatarURL,
		AvatarThumb:       result.Data.AvatarThumb,
		AvatarMiddle:      result.Data.AvatarMiddle,
		AvatarBig:         result.Data.AvatarBig,
		OpenID:            result.Data.OpenID,
		UnionID:           result.Data.UnionID,
		Email:             result.Data.Email,
		EnterpriseEmail:   result.Data.EnterpriseEmail,
		UserID:            result.Data.UserID,
		Mobile:            result.Data.Mobile,
		EmployeeNo:        result.Data.EmployeeNo,
	}, nil
}
