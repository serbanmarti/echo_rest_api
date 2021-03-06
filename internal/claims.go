package internal

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	JWTClaims struct {
		Role string `json:"role"`
		jwt.StandardClaims
	}
)

// DecodeClaims decodes a set of JWT claims
func DecodeClaims(c echo.Context) *JWTClaims {
	user := c.Get("user").(*jwt.Token)
	return user.Claims.(*JWTClaims)
}

// IsAdmin checks if the logged in user has admin rights
func (c *JWTClaims) IsAdmin() error {
	if c.Role != "admin" {
		return NewError(ErrBENotAdmin, nil, 2)
	}

	return nil
}

// GetUserID retrieves the logged in user ID
func (c *JWTClaims) GetUserID() (primitive.ObjectID, error) {
	// Convert the ID into a proper ObjectID
	id, err := primitive.ObjectIDFromHex(c.Id)
	if err != nil {
		return id, NewError(ErrBEMongoIDCast, nil, 2)
	}

	return id, nil
}
