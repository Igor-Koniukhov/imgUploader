package models

type JWTPayload struct {
	Sub             string `json:"sub"`
	EmailVerified   bool   `json:"email_verified"`
	Iss             string `json:"iss"`
	CognitoUsername string `json:"cognito:username"`
	OriginJti       string `json:"origin_jti"`
	Aud             string `json:"aud"`
	EventID         string `json:"event_id"`
	TokenUse        string `json:"token_use"`
	AuthTime        int64  `json:"auth_time"`
	Name            string `json:"name"`
	Exp             int64  `json:"exp"`
	Iat             int64  `json:"iat"`
	Jti             string `json:"jti"`
	Email           string `json:"email"`
}
