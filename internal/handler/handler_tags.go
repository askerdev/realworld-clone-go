package handler

import "net/http"

func (h *handler) listTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.storage.SelectTags(r.Context())
	if err != nil {
		InternalServerError(w)
		return
	}

	JSON(w, map[string]any{
		"tags": tags,
	})
}
