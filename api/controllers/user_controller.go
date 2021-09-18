package controllers

import (
	"context"
	"fmt"
	"log"

	// "strconv"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"link-shortener-api/api/db"
	"link-shortener-api/api/helpers"

	// helper "link-shortener-api/helpers"
	// "link-shortener-api/api/models/entity"

	"link-shortener-api/api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct{}

var userCollection *mongo.Collection = db.OpenCollection(db.Client, "users")
var validate = validator.New()

//HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

//VerifyPassword checks the input password while verifying it with the password in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or password is incorrect"
		check = false
	}

	return check, msg
}

func (uc UserController) SignUp(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErr := validate.Struct(user)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
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

	password := HashPassword(user.Password)
	user.Password = password
	user.IP = c.ClientIP()

	user.Country = helpers.IPInfo(user.IP)
	user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	// user.UserId = user.ID.Hex()
	token, _ := helpers.GenerateToken("USER", user.ID.Hex(), user.Email)
	// user.Token = &token
	// user.Refresh_token = &refreshToken

	result, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		log.Printf("Could not create User: %v", insertErr)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"userId": result.InsertedID, "accessToken": token, "accessTokenDuration": 24})
}

//Login user
func (uc UserController) Login(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User
	var foundUser models.User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
		return
	}

	passwordIsValid, msg := VerifyPassword(user.Password, foundUser.Password)
	if !passwordIsValid {
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	token, _ := helpers.GenerateToken("USER", foundUser.ID.Hex(), foundUser.Email)

	c.JSON(http.StatusOK, gin.H{"userId": foundUser.ID, "accessToken": token, "accessTokenDuration": 24})
}

func (uc UserController) GetAllUsers(c *gin.Context) {
	var u []bson.M

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	lookupStage := bson.D{{"$lookup", bson.D{{"from", "links"}, {"localField", "links"}, {"foreignField", "_id"}, {"as", "links"}}}}
	projectStage := bson.D{
		{"$project", bson.D{
			{"password", 0},
		}}}

	showLoadedStructCursor, err := userCollection.Aggregate(ctx, mongo.Pipeline{lookupStage, projectStage})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = showLoadedStructCursor.All(ctx, &u)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}
