package controllers

import (
	"context"
	"log"
	"moodring-api/database"
	"moodring-api/models"
	utils "moodring-api/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
	validate                         = validator.New()
)

func HashPassword(password string) string {
	// TODO: Change `bytes` to something meaningful
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
    print("oops")
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "email or password is incorrect"
		check = false
	}

	return check, msg
}

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
			print("ohhh")
      log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for email"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

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

		token, refreshToken, _ := utils.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserID)

		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		defer cancel()

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is not correct"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		// TODO: The `_` being returned below is an error. Add handling for it.
		token, refreshToken, _ := utils.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserID)
		utils.UpdateAllTokens(token, refreshToken, foundUser.UserID)

		err = userCollection.FindOne(ctx, bson.M{"userID": foundUser.UserID}).Decode(&foundUser)

    if err != nil{
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }

    c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
  return func(c *gin.Context){
    // Make sure only admin is calling this function 
    if err := utils.CheckUserType(c, "ADMIN"); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

    recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
    if err != nil || recordPerPage < 1{
      recordPerPage = 10
    }

    page, err1 := strconv.Atoi(c.Query("page"))
    if err1 != nil || page<1{
      page = 1
    }

    startIndex := (page - 1) * recordPerPage
    if queryStartIndex := c.Query("startIndex"); queryStartIndex != "" {
      if index, err := strconv.Atoi(queryStartIndex); err == nil {
        startIndex = index
      }
    }

    matchStage := bson.D{
      {Key: "$match", Value: bson.D{}},
    }

    groupStage := bson.D{
      {Key: "$group", Value: bson.D{
        {Key: "_id", Value: nil},  
        {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
        {Key: "all_docs", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
      }},
    }
    
    projectStage := bson.D{
      {Key: "$project", Value: bson.D{
          {Key: "_id", Value: 0},
          {Key: "total_count", Value: 1},
          {Key: "user_items", Value: bson.D{
            {Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
          }},
      }},
    }

    result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})

    defer cancel()
    if err!= nil{
      c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
    }

    var allUsers []bson.M
    if err = result.All(ctx, &allUsers); err != nil {
      log.Fatal(err)
    }

    c.JSON(http.StatusOK, allUsers[0])
  }
}

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
