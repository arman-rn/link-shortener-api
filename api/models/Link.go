package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Link struct {
	ID        primitive.ObjectID `bson:"_id"`
	User      primitive.ObjectID `json:"user"`
	LongUrl   string             `json:"longUrl" bson:"longUrl"`
	ShortUrl  string             `json:"shortUrl,omitempty" bson:"shortUrl,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
