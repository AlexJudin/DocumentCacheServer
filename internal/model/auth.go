package model

type Token struct {
	AccessTokenID string `gorm:"primarykey"`
	Login         string
}
