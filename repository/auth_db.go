package repository

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AuthSqlite struct {
	db *sqlx.DB
}

type User struct {
	Login    string `json:"-"`
	Password string `json:"password"`
}

type myClaims struct {
	jwt.StandardClaims
	Login string `json:"login"`
}

func NewAuthSqlite(db *sqlx.DB) *AuthSqlite {
	return &AuthSqlite{db: db}
}

func (a *AuthSqlite) CheckAuth(c *gin.Context) {
	var u User

	err := c.ShouldBindJSON(&u)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if u.Password == "" {
		c.JSON(401, gin.H{"error": "нет пароля"})
		return
	}

	passwordENV := os.Getenv("TODO_PASSWORD")

	if len(passwordENV) == 0 {
		c.JSON(200, gin.H{"warning": "пароль не задан"})
		return
	}

	hashPassENV := generatePasswordHash(passwordENV)

	if u.Password != "" {
		hashPass := generatePasswordHash(u.Password)

		if hashPass != hashPassENV {
			c.JSON(401, gin.H{"warning": "неверный пароль"})
			return
		}
	}

	if generatePasswordHash(u.Password) == hashPassENV {
		token, err := GenerateJWT(u.Login)
		if err != nil {
			logrus.Error(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"token": token})
	}
}

func GenerateJWT(username string) (string, error) {
	if username == "" {
		username = "default"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &myClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(8 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		username,
	})

	return token.SignedString([]byte(viper.Get("SIGN_KEY").(string)))
}

func generatePasswordHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return fmt.Sprint("%x", hash.Sum(nil))
}
