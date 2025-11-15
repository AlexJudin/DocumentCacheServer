package service

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

const (
	minLoginLength    = 8
	minPasswordLength = 8
)

type AuthService struct {
	Config       *config.Config
	TokenStorage postgres.TokenStorage
}

func NewAuthService(cfg *config.Config, db *gorm.DB) AuthService {
	return AuthService{
		Config:       cfg,
		TokenStorage: postgres.NewTokenStorageRepo(db),
	}
}

func (s AuthService) GenerateHashPassword(password string) string {
	passwordBytes := []byte(password)
	sha512Hasher := sha512.New()

	passwordBytes = append(passwordBytes, s.Config.PasswordSalt...)
	sha512Hasher.Write(passwordBytes)

	hashedPasswordBytes := sha512Hasher.Sum(nil)
	hashedPasswordHex := hex.EncodeToString(hashedPasswordBytes)

	return hashedPasswordHex
}

func (s AuthService) GenerateTokens(login string) (entity.Tokens, error) {
	accessTokenID := uuid.NewString()
	accessToken, err := s.generateAccessToken(login)
	if err != nil {
		return entity.Tokens{}, err
	}

	refreshToken, err := s.generateRefreshToken(login, accessTokenID)
	if err != nil {
		return entity.Tokens{}, err
	}

	token := model.Token{
		AccessTokenID: accessTokenID,
		Login:         login,
	}
	err = s.TokenStorage.Save(token)
	if err != nil {
		return entity.Tokens{}, err
	}

	return entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s AuthService) VerifyUser(token string) (string, error) {
	claims := &entity.AuthClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect method")
		}

		return s.Config.TokenSalt, nil
	})

	if err != nil || !parsedToken.Valid {
		return "", fmt.Errorf("incorrect token: %+v", err)
	}

	return claims.Login, nil
}

func (s AuthService) RefreshToken(refreshToken string) (entity.Tokens, error) {
	claims := &entity.RefreshTokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect method")
		}

		return s.Config.TokenSalt, nil
	})

	if err != nil || !parsedToken.Valid {
		return entity.Tokens{}, fmt.Errorf("incorrect refresh refreshToken: %+v", err)
	}

	// поиск токена в хранилище claims.AccessTokenID
	login, err := s.TokenStorage.Get(claims.AccessTokenID)
	if err != nil || login != claims.Login {
		return entity.Tokens{}, custom_error.ErrUserNotFound
	}

	tokens, err := s.GenerateTokens(claims.Login)
	if err != nil {
		return entity.Tokens{}, err
	}

	// удалить старую пару токенов claims.AccessTokenID
	err = s.TokenStorage.Delete(claims.AccessTokenID)
	if err != nil {
		return entity.Tokens{}, err
	}

	return tokens, nil
}

func (s AuthService) DeleteToken(refreshToken string) error {
	claims := &entity.RefreshTokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect method")
		}

		return s.Config.TokenSalt, nil
	})

	if err != nil || !parsedToken.Valid {
		return fmt.Errorf("incorrect token: %+v", err)
	}

	err = s.TokenStorage.Delete(claims.AccessTokenID)
	if err != nil {
		return err
	}

	return nil
}

func (s AuthService) LoginIsValid(login string) bool {
	if len(login) < minLoginLength {
		return false
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", login)
	if !matched {
		return false
	}

	return true
}

func (s AuthService) PasswordIsValid(password string) bool {
	if len(password) < minPasswordLength {
		return false
	}

	isValid := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			isValid = true
		case unicode.IsLower(char):
			isValid = true
		case unicode.IsDigit(char):
			isValid = true
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			isValid = true
		default:
			isValid = false
		}
	}

	if !isValid {
		return false
	}

	return true
}

func (s AuthService) generateAccessToken(login string) (string, error) {
	now := time.Now()
	claims := entity.AuthClaims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Config.AccessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.Config.TokenSalt) //переделать на []byte
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s AuthService) generateRefreshToken(login string, accessTokenID string) (string, error) {
	now := time.Now()
	claims := entity.RefreshTokenClaims{
		Login:         login,
		AccessTokenID: accessTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Config.RefreshTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.Config.TokenSalt)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
