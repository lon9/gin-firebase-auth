package ginfirebaseauth

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

const valName = "FIREBASE_ID_TOKEN"

// FirebaseAuthMiddleware is middleware for Firebase Authentication
type FirebaseAuthMiddleware struct {
	cli          *auth.Client
	unAuthorized func(c *gin.Context)
}

// New is constructor of the middleware
func New(credFileName string, unAuthorized func(c *gin.Context)) (*FirebaseAuthMiddleware, error) {
	opt := option.WithCredentialsFile(credFileName)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}
	auth, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &FirebaseAuthMiddleware{
		cli:          auth,
		unAuthorized: unAuthorized,
	}, nil
}

// MiddlewareFunc is function to verify token
func (fam *FirebaseAuthMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		token := strings.Replace(authHeader, "Bearer ", "", 1)
		idToken, err := fam.cli.VerifyIDToken(context.Background(), token)
		if err != nil {
			if fam.unAuthorized == nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  http.StatusUnauthorized,
					"message": http.StatusText(http.StatusUnauthorized),
				})
			} else {
				fam.unAuthorized(c)
			}
			return
		}
		c.Set(valName, *idToken)
		c.Next()
	}
}

// ExtractClaims extracts claims
func ExtractClaims(c *gin.Context) *auth.Token {
	idToken, ok := c.Get(valName)
	if !ok {
		return new(auth.Token)
	}
	return idToken.(*auth.Token)
}
