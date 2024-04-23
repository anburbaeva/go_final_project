package handler

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	CookieName         = "token"
	CookieIsEmptyError = "Поле есть, а токен пустой"
	ErrCookieRequest   = "Ошибка в куках: %v"
	ErrorParsedToken   = "Ошибка во время парсинга токена: %v"
)

func (h *Handler) authMiddleware(c *gin.Context) {
	if os.Getenv("TODO_PASSWORD") != "" {
		cookie, err := c.Request.Cookie(CookieName)
		if err != nil {
			logrus.Printf(ErrCookieRequest, err.Error())
			NewResponseError(c, 401, err.Error())
			return
		}

		if cookie.Value == "" {
			logrus.Println(CookieIsEmptyError)
			NewResponseError(c, 401, CookieIsEmptyError)
			return
		}

		_, err = h.service.Authorization.ParseToken(cookie.Value)
		if err != nil {
			logrus.Printf(ErrorParsedToken, err.Error())
			NewResponseError(c, 401, err.Error())
			return
		}
	}
}
