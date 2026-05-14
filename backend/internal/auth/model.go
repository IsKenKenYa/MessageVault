package auth

type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type UserInfo struct {
	Buttons  []string `json:"buttons"`
	Roles    []string `json:"roles"`
	UserID   string   `json:"userId"`
	UserName string   `json:"userName"`
	Email    string   `json:"email"`
}
