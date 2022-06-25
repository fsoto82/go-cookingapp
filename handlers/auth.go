package handlers

import (
	"fmt"
	"github.com/auth0-community/go-auth0"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/square/go-jose.v2"
	"log"
	"net/http"
	"time"
)

type AuthHandler struct {
	auth0Client   *auth0.JWKClient
	configuration auth0.Configuration
	validator     *auth0.JWTValidator
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(auth0Domain, apiId string) *AuthHandler {
	fullDomain := fmt.Sprintf("https://%s/", auth0Domain)
	auth0Client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: fullDomain + ".well-known/jwks.json"}, nil)
	configuration := auth0.NewConfiguration(auth0Client, []string{apiId}, fullDomain, jose.RS256)
	return &AuthHandler{
		auth0Client:   auth0Client,
		configuration: configuration,
		validator:     auth0.NewValidator(configuration, nil),
	}
}

func (handler *AuthHandler) AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := handler.validator.ValidateRequest(c.Request)
		if err != nil {
			log.Println("Invalid token: ", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

/*
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

	token := jwt.NewWithClaims(handler.signingMethod, claims)
	tokenString, err := token.SignedString(handler.jwtSecret)
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
		return []byte(handler.jwtSecret), nil
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
	token := jwt.NewWithClaims(handler.signingMethod, claims)
	tokenString, err := token.SignedString(handler.jwtSecret)
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
*/
