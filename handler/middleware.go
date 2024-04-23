package handler

import (
	"os"

	"github.com/gin-gonic/gin"
)

func (h *Handler) authMiddleware(c *gin.Context) {
	if os.Getenv("TODO_PASSWORD") != "" {
		cookie, err := c.Request.Cookie("token")
		if err != nil {
			NewResponseError(c, 401, err.Error())
			return
		}

		if cookie.Value == "" {
			NewResponseError(c, 401, "поле для токена пустое")
			return
		}

		_, err = h.service.Authorization.ParseToken(cookie.Value)
		if err != nil {
			NewResponseError(c, 401, err.Error())
			return
		}
	}
}
