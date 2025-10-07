package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/api/common"
	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

var messageError string

type AuthHandler struct {
	uc usecases.Authorization
}

func NewAuthHandler(uc usecases.Authorization) AuthHandler {
	return AuthHandler{uc: uc}
}

func (h *AuthHandler) AuthorizationUser(w http.ResponseWriter, r *http.Request) {
	var (
		user model.User
		buf  bytes.Buffer
	)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("authorization user error: %+v", err)
		messageError = "Переданы некорректные логин/пароль."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		log.Errorf("authorization user error: %+v", err)
		messageError = "Не удалось прочитать логин/пароль."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	tokens, err := h.uc.AuthorizationUser(user.Login, user.Password)
	switch {
	case errors.Is(err, custom_error.ErrUserNotFound):
		log.Errorf("authorization user error: %+v", err)
		messageError = "Пользователь не найден."

		common.ApiError(http.StatusNotFound, messageError, w)
		return
	case errors.Is(err, custom_error.ErrIncorrectPassword):
		log.Errorf("authorization user error: %+v", err)
		messageError = "Некорректный пароль."

		common.ApiError(http.StatusForbidden, messageError, w)
		return
	case err != nil:
		log.Errorf("authorization user error: %+v", err)
		messageError = "Ошибка сервера, не удалось авторизовать пользователя. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	accessTokenCookie := http.Cookie{
		Name:     "accessToken",
		Value:    tokens.AccessToken,
		HttpOnly: true,
	}
	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
	}
	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	respMap := entity.ApiResponse{
		Response: map[string]interface{}{
			"token": tokens.AccessToken,
		},
	}

	resp, err := json.Marshal(respMap)
	if err != nil {
		log.Errorf("authorization user error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("authorization user error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		log.Error("refresh token not found")
		messageError = "refresh token не найден"

		common.ApiError(http.StatusUnauthorized, messageError, w)
		return
	}

	tokens, err := h.uc.RefreshToken(refreshToken.Value)
	if err != nil {
		if errors.Is(err, custom_error.ErrUserNotFound) {
			log.Error("token not found in storage")
			messageError = "Токен не найден"

			common.ApiError(http.StatusNotFound, messageError, w)
			return
		}

		log.Error("cannot refresh token")
		messageError = "Не удалось обновить токен"

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	accessTokenCookie := http.Cookie{
		Name:     "accessToken",
		Value:    tokens.AccessToken,
		HttpOnly: true,
	}
	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
	}
	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	accessToken, err := r.Cookie("accessToken")
	if err != nil {
		log.Error("access token not found")
		messageError = "access token не найден"

		common.ApiError(http.StatusUnauthorized, messageError, w)
		return
	}

	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		log.Error("refresh token not found")
		messageError = "refresh token не найден"

		common.ApiError(http.StatusUnauthorized, messageError, w)
		return
	}

	err = h.uc.DeleteToken(refreshToken.Value)
	if err != nil {
		log.Errorf("authorization user error: %+v", err)
		messageError = "Не удалось завершить авторизованную сессию работы"

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	accessTokenCookie := http.Cookie{
		Name:     "accessToken",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	}

	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	response := entity.ApiResponse{
		Response: map[string]interface{}{
			accessToken.Value: true,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("register user error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Errorf("register user error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}
