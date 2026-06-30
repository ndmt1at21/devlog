package handler

import (
	"net/http"
	"strings"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

type commentDTO struct {
	Name    string `json:"name"`
	Text    string `json:"text"`
	Time    string `json:"time"`
	Initial string `json:"initial"`
}

func toCommentDTO(c domain.Comment) commentDTO {
	return commentDTO{Name: c.Name, Text: c.Body, Time: relativeVN(c.CreatedAt), Initial: initial(c.Name)}
}

func (a *API) listComments(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	cs, err := a.Store.Comments().ListByArticle(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Không tải được bình luận.")
		return
	}
	out := make([]commentDTO, 0, len(cs))
	for _, c := range cs {
		out = append(out, toCommentDTO(c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) createComment(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	var in struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Text = strings.TrimSpace(in.Text)
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "Vui lòng nhập tên của bạn.")
		return
	}
	if in.Text == "" {
		writeError(w, http.StatusBadRequest, "Bình luận không được để trống.")
		return
	}

	// Ensure the article exists so comments can't be orphaned.
	if _, err := a.Store.Articles().GetBySlug(r.Context(), slug); err != nil {
		writeError(w, http.StatusNotFound, "Không tìm thấy bài viết.")
		return
	}

	created, err := a.Store.Comments().Create(r.Context(), domain.Comment{
		ArticleSlug: slug, Name: in.Name, Body: in.Text,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Không gửi được bình luận.")
		return
	}
	writeJSON(w, http.StatusCreated, toCommentDTO(created))
}
