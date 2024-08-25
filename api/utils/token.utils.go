package utils

import (
	"context"
	"log"
	"moodring-api/database"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if err != nil {
		log.Panic(err)
		return
	}
 
  refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(API_SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func ValidateToken(signedToken string) (claims *CustomClaims, msg string){
  token, err := jwt.ParseWithClaims(
    signedToken,
    &CustomClaims{},
    func(token *jwt.Token)(interface{}, error){
      return []byte(API_SECRET_KEY), nil
    },
  )

  if err != nil {
    msg = err.Error()
    return
  }

  claims, ok := token.Claims.(*CustomClaims)
  if !ok {
    msg = "the token is invalid";
    msg = err.Error() // TODO: Mayne not needed
    return;
  }

  // TODO: Create seperate `isTokenExpired` function
  if claims.ExpiresAt.Before(time.Now()){
    msg = "token is expired"
    msg = err.Error() // TODO: Maybe note needed
    return  
  }

  return claims, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userID string){
  var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
  
  var updateObj primitive.D
  updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
  updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: signedRefreshToken})

  UpdatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
  updateObj = append(updateObj, bson.E{Key: "updatedAt", Value: UpdatedAt})

  upsert := true
  filter := bson.M{"userID": userID}
  opt := options.UpdateOptions{
    Upsert: &upsert,
  }

  _, err := userCollection.UpdateOne(
    ctx,
    filter,
    bson.D{
      {Key: "$set", Value: updateObj},
    },
    &opt,
  )

  defer cancel()

  if err != nil {
    log.Panic(err)
    return
  }

  return
} 


