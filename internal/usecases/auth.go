package usecases

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/service"
)

var _ Authorization = (*AuthUsecase)(nil)

type AuthUsecase struct {
	UserDB      postgres.User
	ServiceAuth service.AuthService
}

func NewAuthUsecase(db postgres.User, serviceAuth service.AuthService) *AuthUsecase {
	return &AuthUsecase{
		UserDB:      db,
		ServiceAuth: serviceAuth,
	}
}

func (u *AuthUsecase) AuthorizationUser(login string, password string) (entity.Tokens, error) {
	user, err := u.UserDB.GetByLogin(login)
	if err != nil {
		return entity.Tokens{}, err
	}

	if user.IsNotFound() {
		return entity.Tokens{}, custom_error.ErrUserNotFound
	}

	passwordHash := u.ServiceAuth.GenerateHashPassword(password)
	if user.Hash != passwordHash {
		return entity.Tokens{}, custom_error.ErrIncorrectPassword
	}

	return u.ServiceAuth.GenerateTokens(login)
}

func (u *AuthUsecase) RefreshToken(refreshToken string) (entity.Tokens, error) {
	return u.ServiceAuth.RefreshToken(refreshToken)
}

func (u *AuthUsecase) DeleteToken(refreshToken string) error {
	return u.ServiceAuth.DeleteToken(refreshToken)
}
