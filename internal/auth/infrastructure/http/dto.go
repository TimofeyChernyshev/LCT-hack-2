package http

type registerRequest struct {
	Email                     string `json:"email"`
	Password                  string `json:"password"`
	Role                      string `json:"role"`
	ConsentPdn                bool   `json:"consentPdn"`
	ConsentProfilePublication bool   `json:"consentProfilePublication"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenRequest struct {
	Token string `json:"token"`
}

type emailRequest struct {
	Email string `json:"email"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type deleteMeRequest struct {
	Password string `json:"password"`
}

// ---------- responses ----------

type userResponse struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	Role            string  `json:"role"`
	Status          string  `json:"status"`
	EmailVerifiedAt *string `json:"emailVerifiedAt,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}

type tokenPairResponse struct {
	AccessToken      string       `json:"accessToken"`
	RefreshToken     string       `json:"refreshToken"`
	AccessExpiresAt  string       `json:"accessExpiresAt"`
	RefreshExpiresAt string       `json:"refreshExpiresAt"`
	User             userResponse `json:"user"`
}

type consentResponse struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Version   string  `json:"version"`
	GrantedAt string  `json:"grantedAt"`
	RevokedAt *string `json:"revokedAt,omitempty"`
}

type registerResponse struct {
	UserID string `json:"userId"`
}
