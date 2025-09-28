package register

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/api/common"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

var messageError string

type RegisterHandler struct {
	uc usecases.Register
}

func NewRegisterHandler(uc usecases.Register) RegisterHandler {
	return RegisterHandler{uc: uc}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var (
		user model.User
		buf  bytes.Buffer
	)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("register user error: %+v", err)
		messageError = "Переданы некорректные логин/пароль."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		log.Errorf("register user error: %+v", err)
		messageError = "Не удалось прочитать логин/пароль."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	err = h.uc.RegisterUser(user.AdminToken, user.Login, user.Password)
	switch {
	case errors.Is(err, custom_error.ErrInvalidAdminToken):
		log.Errorf("register user error: %+v", err)
		messageError = "У пользователя нет прав для регистрации. Обратитесь к администратору."

		common.ApiError(http.StatusForbidden, messageError, w)
		return
	case errors.Is(err, custom_error.ErrInvalidLogin):
		log.Errorf("register user error: %+v", err)
		messageError = "Логин не соответствует требованиям: минимальная длина 8, латиница и цифры."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	case errors.Is(err, custom_error.ErrInvalidPassword):
		log.Errorf("register user error: %+v", err)
		messageError = "Пароль не соответствует требованиям: минимальная длина 8, минимум 2 буквы в разных регистрах, минимум 1 цифра, минимум 1 символ."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	case errors.Is(err, custom_error.ErrUserAlreadyExists):
		log.Errorf("register user error: %+v", err)
		messageError = "Пользователь уже зарегистрирован."

		common.ApiError(http.StatusConflict, messageError, w)
		return
	case err != nil:
		log.Errorf("register user error: %+v", err)
		messageError = "Ошибка сервера, не удалось зарегистрировать пользователя. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	response := entity.ApiResponse{
		Response: map[string]interface{}{
			"login": user.Login,
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
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Errorf("register user error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}
