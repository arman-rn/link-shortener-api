package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Admin struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `json:"name" validate:"required,min=2,max=100"`
	Email     string             `json:"email" validate:"email,required"`
	Password  string             `json:"password" validate:"required,min=6"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
