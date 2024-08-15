package controllers

import (
	"context"
	"fmt"
	"log"
	"moodring-api/database"
	"moodring-api/models"
	utils "moodring-api/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
	validate                         = validator.New()
)

func HashPassword()

func VerifyPassword()

// TODO: This function can be broken up into smaller functions ("doesUserAlreadyExist()")
// Registers a new user in the MongoDB cluster.
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Write function/constants for timeout seconds
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if validationErr := validate.Struct(user); validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// Check email already exists
		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for email"})
		}

		// Check userID already exists
		userIDCount, err := userCollection.CountDocuments(ctx, bson.M{"userID": user.UserID})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for userID"})
			return
		}

		// Throw error if emailCount or userIDCount already in use
		if emailCount > 0 || userIDCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email or userID already in use"})
			return
		}

		// TODO: Using a package for getting time instead of "clunky" RFC3339
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		token, refreshToken, _ := utils.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, *&user.UserID)

		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		defer cancel()

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login()

func GetUsers()

// Retrieves a user from the MongoDB cluster.
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")

		if err := utils.MatchUserTypeToUserID(c, userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"userID": userID}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
