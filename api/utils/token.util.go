package utils

import (
	"log"
	"moodring-api/database"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Move to a seperate types file
type CustomClaims struct {
	Email     string
	FirstName string
	LastName  string
	UserID    string
	UserType  string
	jwt.RegisteredClaims
}

// TODO: Add colletion name as constant
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var API_SECRET_KEY string = os.Getenv("MOODRING_API_SECRET_KEY")

// TODO: Rewrite as general `generateToken` function that accepts the duration of expiration time as a param
func GenerateAllTokens(firstName string, lastName string, email string, userType string, userID string) (signedToken string, signedRefreshToken string, err error) {
	claims := &CustomClaims{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		UserType:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	refreshClaims := &CustomClaims{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		UserType:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 168)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(API_SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(API_SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}
