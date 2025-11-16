package entity

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

const DefaultMimeType = "application/json"

type ApiError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type ApiResponse struct {
	Error    *ApiError              `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type DocumentListRequest struct {
	Login string `json:"login"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Limit int    `json:"limit"`
}

func (d *DocumentListRequest) LoginIsEmpty() bool {
	return d.Login == ""
}

type DocumentFile struct {
	Name    string
	Content []byte
}

type Document struct {
	Meta *model.MetaDocument
	Json map[string]interface{}
	File *DocumentFile
}
