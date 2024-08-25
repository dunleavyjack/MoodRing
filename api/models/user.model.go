package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	FirstName    *string            `json:"firstName" validate:"required,min=2,max=100"`
	LastName     *string            `json:"lastName" validate:"required,min=2,max=100"`
	Password     *string            `json:"password" validate:"required,min=8"`
	Email        *string            `json:"email" validate:"required,email"`
	UserType     *string            `json:"userType" validate:"required,eq=ADMIN|eq=USER"`
	Token        *string            `json:"token"`
	RefreshToken *string            `json:"refreshToken"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
	UserID       string             `json:"userID"`
	ID           primitive.ObjectID `bson:"_id"`
}
