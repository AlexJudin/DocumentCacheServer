package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/api/common"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/entity"
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
	err = json.Unmarshal([]byte(jsonDoc), &jsonDocMap)
	if err != nil {
		log.Errorf("save document error: %+v", err)
		messageError = "Не удалось загрузить json документа."

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
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

		document.File = &entity.DocumentFile{
			Name:    header.Filename,
			Content: file,
		}
	}
	document.Json = jsonDocMap

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
			"file": document.Meta.FilePath,
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
		w.Header().Set("Content-Type", "application/json")
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
	if err != nil {
		log.Errorf("get document by id error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось получить документ [%s]. Попробуйте позже или обратитесь в тех. поддержку.", idDoc)

		common.ApiError(http.StatusBadRequest, messageError, w)
		return
	}

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "application/json")
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
	if err != nil {
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
	login, ok := r.Context().Value("currentUser").(string)
	if !ok {
		return "", fmt.Errorf("current user not found")
	}

	return login, nil
}
