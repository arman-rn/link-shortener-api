package controllers

import (
	"context"
	"link-shortener-api/api/db"
	"link-shortener-api/api/models"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrorResponse is interface for sending error message with code.
type ErrorResponse struct {
	Code    int
	Message string
}

type LinkController struct{}

var linksCollection *mongo.Collection = db.OpenCollection(db.Client, "links")

// Helper function to handle the HTTP response
func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// GetShortURLHandler This function will return the response based ono user found in Database
func (lc LinkController) URLShortener(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var link models.Link

	type URLRequestObject struct {
		URL string `json:"url"`
	}

	var uro URLRequestObject

	type URLCollection struct {
		ActualURL string
		ShortURL  string
	}
	type SuccessResponse struct {
		Code     int
		Message  string
		Response URLCollection
	}

	err := c.BindJSON(&uro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "URL can't be empty"})
	} else if !isURL(uro.URL) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An invalid URL found, provide a valid URL"})
	} else {
		uniqueID, idError := shortid.Generate()
		if idError != nil {
			log.Println(idError)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while processing"})
		} else {
			link.ShortUrl = uniqueID
			link.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			link.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			link.ID = primitive.NewObjectID()
			link.LongUrl = uro.URL

			user, _ := c.Get("user")
			userId := user.(models.User).ID

			link.User = userId

			result, insertErr := linksCollection.InsertOne(ctx, link)
			if insertErr != nil {
				log.Printf("Could not create Link: %v", insertErr)

				c.JSON(http.StatusInternalServerError, gin.H{"error": "Link item was not created"})
				return
			}

			update := bson.M{
				"$push": bson.M{"links": result.InsertedID},
			}

			userCollection.FindOneAndUpdate(ctx, bson.M{"_id": userId}, update)

			var successResponse = SuccessResponse{
				Code:    http.StatusOK,
				Message: "Short URL generated",
				Response: URLCollection{
					ActualURL: uro.URL,
					ShortURL:  c.Request.Host + "/" + uniqueID,
				},
			}
			c.JSON(http.StatusOK, successResponse)
		}
	}
}

func (lc LinkController) ShortURLHandler(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var foundLink models.Link

	url := c.Param("url")
	if url == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "URL Code can't be empty"})
		return
	}

	err := linksCollection.FindOne(ctx, bson.M{"shortUrl": url}).Decode(&foundLink)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid url"})
		return
	}

	c.Redirect(http.StatusSeeOther, foundLink.LongUrl)
}
