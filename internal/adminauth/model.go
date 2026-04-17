package adminauth

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// AdminClaimType is the value stored in the JWT "typ" custom claim to mark an admin token.
const AdminClaimType = "admin"
