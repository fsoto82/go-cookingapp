package handlers

import (
	"cookingapp/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"time"
)

type AuthHandler struct {
	JwtSecret     string
	SigningMethod *jwt.SigningMethodHMAC
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(jwtSecret string, signingMethod *jwt.SigningMethodHMAC) *AuthHandler {
	return &AuthHandler{JwtSecret: jwtSecret, SigningMethod: signingMethod}
}

func (handler *AuthHandler) AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(handler.JwtSecret), nil
		})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if tkn == nil || !tkn.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}

func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error reading data: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading data"})
		return
	}
	if user.Username != "admin" || user.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}

	expirationTime := time.Now().Add(10 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: expirationTime},
		},
	}

	token := jwt.NewWithClaims(handler.SigningMethod, claims)
	tokenString, err := token.SignedString(handler.JwtSecret)
	if err != nil {
		log.Println("Error tokenizing: ", err)
		c.Status(http.StatusBadRequest)
		return
	}

	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOutput)
}

func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(handler.JwtSecret), nil
	})
	if err != nil {
		log.Println("Error parsing claims: ", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if tkn == nil || !tkn.Valid {
		log.Println("Invalid token: ", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		log.Println("Token is not expired yet")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = &jwt.NumericDate{Time: expirationTime}
	token := jwt.NewWithClaims(handler.SigningMethod, claims)
	tokenString, err := token.SignedString(handler.JwtSecret)
	if err != nil {
		log.Println("Error tokenizing: ", err)
		c.Status(http.StatusBadRequest)
		return
	}

	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOutput)
}
