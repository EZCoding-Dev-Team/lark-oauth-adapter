package dto

type UserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	EnName            string `json:"en_name"`
	AvatarURL         string `json:"avatar_url"`
	AvatarThumb       string `json:"avatar_thumb"`
	AvatarMiddle      string `json:"avatar_middle"`
	AvatarBig         string `json:"avatar_big"`
	OpenID            string `json:"open_id"`
	UnionID           string `json:"union_id"`
	Email             string `json:"email"`
	EnterpriseEmail   string `json:"enterprise_email"`
	UserID            string `json:"user_id"`
	Mobile            string `json:"mobile"`
	EmployeeNo        string `json:"employee_no"`
}
