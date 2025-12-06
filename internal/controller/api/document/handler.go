package document

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/controller/common"
	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

var messageError string

type DocumentHandler struct {
	uc usecases.Document
}

func NewDocumentHandler(uc usecases.Document) DocumentHandler {
	return DocumentHandler{
		uc: uc,
	}
}

// SaveDocument godoc
// @Summary Сохранить документ
// @Description Загружает и сохраняет документ с метаданными и опциональным JSON содержимым
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param meta formData string true "Метаданные документа в формате JSON"
// @Param json formData string false "JSON содержимое документа"
// @Param file formData file false "Файл документа (если meta.file = true)"
// @Success 201 {object} entity.ApiResponse "Документ успешно сохранен"
// @Failure 400 {object} entity.ApiError "Некорректные параметры запроса"
// @Failure 500 {object} entity.ApiError "Внутренняя ошибка сервера"
// @Failure 503 {object} entity.ApiError "Сервер недоступен"
// @Router /docs [post]
func (h *DocumentHandler) SaveDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Errorf("save document error: %+v", err)
		messageError = "Переданы некорректные параметры запроса на загрузку нового документа."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	metaDoc := r.FormValue("meta")

	var document entity.Document

	if err = json.Unmarshal([]byte(metaDoc), &document.Meta); err != nil {
		log.Errorf("save document error: %+v", err)
		messageError = "Не удалось прочитать параметры документа."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	jsonDoc := r.FormValue("json")

	jsonDocMap := make(map[string]interface{})
	if jsonDoc != "" {
		err = json.Unmarshal([]byte(jsonDoc), &jsonDocMap)
		if err != nil {
			log.Errorf("save document error: %+v", err)
			messageError = "Не удалось загрузить json документа."

			common.ApiError(http.StatusBadRequest, messageError, w)
			return
		}

		document.Json = jsonDocMap
	}

	if document.Meta.File {
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Errorf("save document error: %+v", err)
			messageError = "Не удалось загрузить файл документа."

			common.ApiError(http.StatusBadRequest, messageError, w)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			log.Errorf("save document error: %+v", err)
			messageError = "Не удалось прочитать файл документа."

			common.ApiError(http.StatusBadRequest, messageError, w)
			return
		}

		document.File = &entity.DocumentFile{
			Name:    header.Filename,
			Content: data,
		}
	}

	err = h.uc.SaveDocument(&document)
	if err != nil {
		log.Errorf("save document: error save document [%s]: service is not allowed", document.Meta.Name)
		messageError = "Ошибка сервера, не удалось сохранить документ. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	respMap := entity.ApiResponse{
		Data: map[string]interface{}{
			"json": jsonDocMap,
			"file": document.Meta.Name,
		},
	}

	resp, err := json.Marshal(respMap)
	if err != nil {
		log.Errorf("save document error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("save document error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}

// GetDocumentsList godoc
// @Summary Получить список документов
// @Description Возвращает список документов с возможностью фильтрации по пользователю
// @Tags documents
// @Accept json
// @Produce json
// @Param request body entity.DocumentListRequest true "Параметры запроса списка документов"
// @Success 200 {object} entity.ApiResponse "Список документов успешно получен"
// @Success 200 {object} nil "Для HEAD запроса - только проверка доступности"
// @Failure 400 {object} entity.ApiError "Некорректные параметры запроса"
// @Failure 500 {object} entity.ApiError "Внутренняя ошибка сервера"
// @Failure 503 {object} entity.ApiError "Сервер недоступен"
// @Router /docs [get]
// @Router /docs [head]
func (h *DocumentHandler) GetDocumentsList(w http.ResponseWriter, r *http.Request) {
	var (
		req entity.DocumentListRequest
		buf bytes.Buffer
	)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("get documents list error: %+v", err)
		messageError = "Переданы некорректные параметры для получения списка документов."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &req); err != nil {
		log.Errorf("get documents list error: %+v", err)
		messageError = "Не удалось прочитать параметры для получения списка документов."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	if req.LoginIsEmpty() {
		login, err := getCurrentUser(r)
		if err != nil {
			log.Errorf("get documents list error: %+v", err)
			messageError = "Внутренняя ошибка сервера. Не удалось получить логин текущего пользователя."

			common.ApiError(http.StatusInternalServerError, messageError, w)
			return
		}

		req.Login = login
	}

	documentList, err := h.uc.GetDocumentsList(req)
	if err != nil {
		log.Error("get documents list error: service is not allowed")
		messageError = "Ошибка сервера, не удалось получить список документов. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
	}

	respMap := entity.ApiResponse{
		Data: map[string]interface{}{
			"docs": documentList,
		},
	}

	resp, err := json.Marshal(respMap)
	if err != nil {
		log.Errorf("get documents list error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("get documents list error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}

// GetDocumentById godoc
// @Summary Получить документ по ID
// @Description Возвращает документ по его идентификатору
// @Tags documents
// @Accept json
// @Produce octet-stream
// @Produce json
// @Param id query string true "Идентификатор документа"
// @Success 200 {file} byte "Документ успешно получен"
// @Success 200 {object} nil "Для HEAD запроса - только проверка доступности"
// @Failure 400 {object} entity.ApiError "Не передан идентификатор документа"
// @Failure 400 {object} entity.ApiError "Документ не найден или ошибка сервера"
// @Failure 503 {object} entity.ApiError "Сервер недоступен"
// @Router /docs/ [get]
// @Router /docs/ [head]
func (h *DocumentHandler) GetDocumentById(w http.ResponseWriter, r *http.Request) {
	idDoc := r.FormValue("id")
	if idDoc == "" {
		err := fmt.Errorf("document id is empty")
		log.Errorf("get document by id error: %+v", err)
		messageError = "Не передан идентификатор документа."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	resp, mime, err := h.uc.GetDocumentById(idDoc)
	switch {
	case errors.Is(err, custom_error.ErrDocumentNotFound):
		log.Errorf("get document by id error: %+v", err)
		messageError = fmt.Sprintf("Документ [%s] не найден.", idDoc)

		common.ApiError(http.StatusNotFound, messageError, w)
		return
	case err != nil:
		log.Errorf("get document by id error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось получить документ [%s]. Попробуйте позже или обратитесь в тех. поддержку.", idDoc)

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", mime)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("get document by id error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}

// DeleteDocumentById godoc
// @Summary Удалить документ по ID
// @Description Удаляет документ по его идентификатору
// @Tags documents
// @Accept json
// @Produce json
// @Param id query string true "Идентификатор документа"
// @Success 200 {object} entity.ApiResponse "Документ успешно удален"
// @Failure 400 {object} entity.ApiError "Не передан идентификатор документа"
// @Failure 400 {object} entity.ApiError "Ошибка при удалении документа"
// @Failure 500 {object} entity.ApiError "Внутренняя ошибка сервера"
// @Failure 503 {object} entity.ApiError "Сервер недоступен"
// @Router /docs/ [delete]
func (h *DocumentHandler) DeleteDocumentById(w http.ResponseWriter, r *http.Request) {
	idDoc := r.FormValue("id")
	if idDoc == "" {
		err := fmt.Errorf("document id is empty")
		log.Errorf("delete document by id error: %+v", err)
		messageError = "Не передан идентификатор документа."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	err := h.uc.DeleteDocumentById(idDoc)
	switch {
	case errors.Is(err, custom_error.ErrDocumentNotFound):
		log.Errorf("delete document by id error: %+v", err)
		messageError = fmt.Sprintf("Документ [%s] не найден.", idDoc)

		common.ApiError(http.StatusNotFound, messageError, w)
		return
	case err != nil:
		log.Errorf("delete document by id error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось удалить документ [%s]. Попробуйте позже или обратитесь в тех. поддержку.", idDoc)

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	respMap := entity.ApiResponse{
		Response: map[string]interface{}{
			idDoc: true,
		},
	}

	resp, err := json.Marshal(respMap)
	if err != nil {
		log.Errorf("delete document by id error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("delete document by id error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		common.ApiError(http.StatusServiceUnavailable, messageError, w)
	}
}

func getCurrentUser(r *http.Request) (string, error) {
	login, ok := r.Context().Value(entity.CurrentUserKey).(string)
	if !ok {
		return "", fmt.Errorf("current user not found")
	}

	return login, nil
}
