package controllers

import (
	"context"
	"fmt"
	"log"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"link-shortener-api/api/db"
	"link-shortener-api/api/helpers"

	"link-shortener-api/api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
)

type AdminController struct{}

var adminCollection *mongo.Collection = db.OpenCollection(db.Client, "admins")

func (uc AdminController) SignUp(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var admin models.Admin

	if err := c.BindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErr := validate.Struct(admin)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	count, err := adminCollection.CountDocuments(ctx, bson.M{"email": admin.Email})
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the email"})
		return
	}

	if count > 0 {
		fmt.Println("this email already exists")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
		return
	}

	password := HashPassword(admin.Password)
	admin.Password = password

	admin.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	admin.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	admin.ID = primitive.NewObjectID()
	token, _ := helpers.GenerateToken("ADMIN", admin.ID.Hex(), admin.Email)

	result, insertErr := adminCollection.InsertOne(ctx, admin)
	if insertErr != nil {
		log.Printf("Could not create Admin: %v", insertErr)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin item was not created"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"adminId": result.InsertedID, "accessToken": token, "accessTokenDuration": 24})
}

//Login admin
func (uc AdminController) Login(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var admin models.Admin
	var foundAdmin models.Admin

	if err := c.BindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := adminCollection.FindOne(ctx, bson.M{"email": admin.Email}).Decode(&foundAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
		return
	}

	passwordIsValid, msg := VerifyPassword(admin.Password, foundAdmin.Password)
	if !passwordIsValid {
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	token, _ := helpers.GenerateToken("ADMIN", foundAdmin.ID.Hex(), foundAdmin.Email)

	c.JSON(http.StatusOK, gin.H{"adminId": foundAdmin.ID, "accessToken": token, "accessTokenDuration": 24})
}
