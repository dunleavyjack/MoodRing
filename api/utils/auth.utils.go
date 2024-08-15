package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// Ensures a user type matches the user id.
// TODO: Rename to checkUserIsAuthorized
func MatchUserTypeToUserID(c *gin.Context, userID string) (err error) {
	userType := c.GetString("userType")
	uid := c.GetString("uid")
	err = nil

	// TODO: Add constant for userTypes
	if userType == "USER" && uid != userID {
		// TODO: Constants for error names
		err = errors.New("unauthorized to access this resource")
		return err
	}

	err = CheckUserType(c, userType)
	return err
}

// Checks to make sure a userType matches the user role.
// TODO: Rename to getUserType
func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("userType")
	err = nil

	if userType != role {
		// TODO Constants for error names
		err = errors.New("unauthorized to access this resource")
		return err
	}

	return err
}
