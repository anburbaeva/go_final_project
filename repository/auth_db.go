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

const (
	TokenTtl               = 8 * time.Hour
	Response               = "Получили объект User со следующими данными: login: %s, password: %s"
	EmptyUserPassword      = "Пользователь передал пустое поле пароля"
	EmptyUserPasswordError = "А где пароль-то?!"
	EnvPassword            = "TODO_PASSWORD"
	NoPasswordInEnv        = "Пароль не задан. Проверь указал ли ты TODO_PASSWORD"
	NoPasswordInEnvLog     = "Пароль не задан. Проверь указал ли ты TODO_PASSWORD в окружении на сверере. Пускаю без пароля"
	WrongPassword          = "Неправильный пароль"
	TokenDone              = "Все проверки прошли. Токен выдали"
	SuccessLogin           = "Похоже, что всё верно - выдаю токен =)"
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
		logrus.Printf(Response, u.Login, u.Password)
		c.JSON(500, gin.H{"error": err})
		return
	}

	if u.Password == "" {
		logrus.Error(EmptyUserPassword)
		c.JSON(401, gin.H{"error": EmptyUserPasswordError})
		return
	}

	passwordENV := os.Getenv(EnvPassword)

	if len(passwordENV) == 0 {
		logrus.Warn(NoPasswordInEnv)
		c.JSON(200, gin.H{"warning": NoPasswordInEnvLog})
		return
	}

	hashPassENV := generatePasswordHash(passwordENV)

	if u.Password != "" {
		hashPass := generatePasswordHash(u.Password)

		if hashPass != hashPassENV {
			logrus.Error(WrongPassword)
			c.JSON(401, gin.H{"warning": WrongPassword})
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
		logrus.Println(TokenDone)
	}
}

func GenerateJWT(username string) (string, error) {
	if username == "" {
		username = "default"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &myClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenTtl).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		username,
	})
	logrus.Println(SuccessLogin)

	return token.SignedString([]byte(viper.Get("SIGN_KEY").(string)))
}

func generatePasswordHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return fmt.Sprint("%x", hash.Sum(nil))
}
