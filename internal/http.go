package internal

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Decode an input parameter into a MongoDB ObjectID
func DecodeParameterID(c echo.Context, p string) (primitive.ObjectID, error) {
	var id primitive.ObjectID

	id, err := primitive.ObjectIDFromHex(c.Param(p))
	if err != nil {
		return id, NewError(ErrBEMongoIDCast, err, 2)
	}

	return id, nil
}

// Decode an input query parameter into a MongoDB ObjectID
func DecodeQueryParameterID(q url.Values, p string) (primitive.ObjectID, error) {
	var id primitive.ObjectID

	rawID := q.Get(p)
	if rawID == "" {
		return id, NewError(ErrBEMongoIDEmpty, nil, 2)
	}

	id, err := primitive.ObjectIDFromHex(rawID)
	if err != nil {
		return id, NewError(ErrBEMongoIDCast, err, 2)
	}

	return id, nil
}
