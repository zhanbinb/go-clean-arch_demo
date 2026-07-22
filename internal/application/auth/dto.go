package auth

// LoginInput is the payload for POST /auth/login.
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=1"`
}

// LoginResult is the response on successful login.
type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// RefreshInput is the payload for POST /auth/refresh.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterInput is the payload for creating a new user.
type RegisterInput struct {
	Username string `json:"username" binding:"required,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}
