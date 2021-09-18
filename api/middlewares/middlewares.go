package middlewares

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"link-shortener-api/api/db"
	"link-shortener-api/api/helpers"
	"link-shortener-api/api/models"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)

		log.Print(string(body))

		c.Next()
	}
}

// Auth validates token and authorizes users
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "No Authorization header provided"})
			return
		}

		temp := strings.Split(authHeader, "Bearer")

		if len(temp) < 2 {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid token"})
			return
		}

		tokenString := strings.TrimSpace(temp[1])

		claims, err := helpers.ValidateToken("USER", tokenString)
		if err != "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var userCollection *mongo.Collection = db.OpenCollection(db.Client, "users")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var opt options.FindOneOptions

		opt.SetProjection(bson.M{"_id": 1, "name": 1, "email": 1})
		objID, _ := primitive.ObjectIDFromHex(claims.ID)

		findError := userCollection.FindOne(ctx, bson.M{"_id": objID, "email": claims.Email}, &opt).Decode(&user)

		if findError != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		c.Set("user", user)

		c.Next()
	}
}

// Auth validates token and authorizes admins
// In case we needed to add some more logic to admin auth I created a separate function
func AdminAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "No Authorization header provided"})
			return
		}

		temp := strings.Split(authHeader, "Bearer")

		if len(temp) < 2 {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid token"})
			return
		}

		tokenString := strings.TrimSpace(temp[1])

		claims, err := helpers.ValidateToken("ADMIN", tokenString)
		if err != "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var adminCollection *mongo.Collection = db.OpenCollection(db.Client, "admins")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var admin models.Admin
		var opt options.FindOneOptions

		opt.SetProjection(bson.M{"_id": 1, "name": 1, "email": 1})
		objID, _ := primitive.ObjectIDFromHex(claims.ID)

		findError := adminCollection.FindOne(ctx, bson.M{"_id": objID, "email": claims.Email}, &opt).Decode(&admin)

		if findError != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		c.Set("admin", admin)

		c.Next()
	}
}
