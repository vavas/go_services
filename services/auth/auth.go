package auth

// Auth structure for services that is passed from auth service to other services.
type Auth struct {
	PlainToken string         `json:"token"`
	User       *User          `json:"user,omitempty"`
	Admin      *Administrator `json:"administrator,omitempty"`
	UserToken  *Token         `json:"user_token,omitempty"`
	AdminToken *Token         `json:"admin_token,omitempty"`
}

// IsAdmin returns true if the request is using an account token or comes from a user with admin permission.
func (auth *Auth) IsAdmin() bool {

	if auth.UserToken == nil || auth.User == nil {
		return false
	}

	return auth.IsSiteAdminToken()
}

// IsSiteAdminToken returns true if the request is using site admin token.
func (auth *Auth) IsSiteAdminToken() bool {
	return auth.isUseSiteAdminToken()
}

func (auth *Auth) isUseSiteAdminToken() bool {
	return auth.UserToken != nil &&
		len(auth.UserToken.Scope) > 0 &&
		auth.UserToken.Scope[0] == "site_admin" &&
		auth.User != nil &&
		auth.User.IsSiteAdmin
}
