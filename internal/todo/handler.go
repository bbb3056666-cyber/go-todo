package todo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	title, err := decodeTitle(r.Body)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	todo, err := h.service.Create(title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			writeResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
		log.Printf("%v", fmt.Errorf("处理创建待办请求失败: %w", err))
		writeResponse(w, http.StatusInternalServerError, "创建待办失败", nil)
		return
	}
	writeResponse(w, http.StatusCreated, "ok", todo)
}

func decodeTitle(body io.Reader) (string, error) {
	decoder := json.NewDecoder(body)
	var fields map[string]json.RawMessage
	if err := decoder.Decode(&fields); err != nil {
		return "", fmt.Errorf("解析请求 JSON 对象失败: %w", err)
	}
	if fields == nil {
		return "", fmt.Errorf("解析请求失败: %w", errors.New("请求体必须为 JSON 对象"))
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err != nil {
			return "", fmt.Errorf("解析请求尾部失败: %w", err)
		}
		return "", fmt.Errorf("解析请求失败: %w", errors.New("请求体只能包含一个 JSON 对象"))
	}
	for field := range fields {
		if field != "title" {
			return "", fmt.Errorf("校验请求字段失败: %w", fmt.Errorf("未知字段 %q", field))
		}
	}
	raw, ok := fields["title"]
	if !ok {
		return "", fmt.Errorf("校验请求失败: %w", errors.New("缺少 title 字段"))
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", fmt.Errorf("校验请求失败: %w", errors.New("title 必须为字符串，不能为 null"))
	}
	var title string
	if err := json.Unmarshal(raw, &title); err != nil {
		return "", fmt.Errorf("解析 title 字符串失败: %w", err)
	}
	return title, nil
}

func writeResponse(w http.ResponseWriter, status int, msg string, data any) {
	code := status
	if status == http.StatusCreated {
		code = 0
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}{Code: code, Msg: msg, Data: data}); err != nil {
		log.Printf("%v", fmt.Errorf("写入 JSON 响应失败: %w", err))
	}
}
