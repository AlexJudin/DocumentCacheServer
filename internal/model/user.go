package model

import "time"

type User struct {
	ID         uint      `gorm:"primarykey" json:"-"`
	CreatedAt  time.Time `json:"-"`
	AdminToken string    `gorm:"-" json:"token"`
	Login      string    `json:"login"`
	Password   string    `gorm:"-" json:"pswd"`
	Hash       string    `json:"-"`
}

func (u User) IsAlreadyExist() bool {
	return u.ID != 0
}

func (u User) IsNotFound() bool {
	return u.ID == 0
}
