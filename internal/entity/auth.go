package entity

type contextKey string

const CurrentUserKey contextKey = "currentUser"

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
