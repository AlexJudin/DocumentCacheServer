package usecases

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/service"
)

var _ Register = (*RegisterUsecase)(nil)

type RegisterUsecase struct {
	DB          postgres.User
	ServiceAuth service.AuthService
}

func NewRegisterUsecase(db postgres.User, serviceAuth service.AuthService) *RegisterUsecase {
	return &RegisterUsecase{
		DB:          db,
		ServiceAuth: serviceAuth,
	}
}

func (u *RegisterUsecase) RegisterUser(token, login, password string) error {
	if token != u.ServiceAuth.Config.AdminToken {
		return custom_error.ErrInvalidAdminToken
	}

	isValid := u.ServiceAuth.LoginIsValid(login)
	if !isValid {
		return custom_error.ErrInvalidLogin
	}

	isValid = u.ServiceAuth.PasswordIsValid(password)
	if !isValid {
		return custom_error.ErrInvalidPassword
	}

	user, err := u.DB.GetUserByLogin(login)
	if err != nil {
		return err
	}

	if user.IsAlreadyExist() {
		return custom_error.ErrUserAlreadyExists
	}

	newUser := model.User{
		Login: login,
		Hash:  u.ServiceAuth.GenerateHashPassword(password),
	}

	err = u.DB.SaveUser(newUser)
	if err != nil {
		return err
	}

	return nil
}
