package database

import (
	"context"

	"echo_rest_api/pkg/internal"
	"echo_rest_api/pkg/model"
	"echo_rest_api/pkg/security"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewDBClientAndConnection create new MongoDB client and connection objects
func NewDBClientAndConnection(ctx context.Context, env *internal.Environ, logger echo.Logger) (*mongo.Client, *mongo.Database) {
	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI(env.DBUri))
	if err != nil {
		logger.Fatal(err)
	}

	// Check that the database connection has been established
	err = dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Fatal("could not establish a DB connection in a timely manner")
	}

	// Create the DB connection pipe
	dbConn := dbClient.Database(env.DBName)

	// Check for DB root account existence; create it if it does not exist
	rootExists, err := model.UserFindRoot(dbConn, env.DBRootUser)
	if err != nil {
		logger.Fatal(err)
	}
	if !rootExists {
		// Generate a new random salt
		s, err := security.NewSalt()
		if err != nil {
			logger.Fatal(err)
		}

		// Hash and set the password field
		p := security.HashPassword(env.DBRootPass, s)

		// Create the new root user object
		u := &model.User{
			Email:        env.DBRootUser,
			Password:     p,
			Salt:         s,
			Role:         "admin",
			Active:       true,
			CreatedUsers: []primitive.ObjectID{},
			InviteToken:  "",
		}

		// Create the user in the DB
		err = model.UserCreateRoot(dbConn, u)
		if err != nil {
			logger.Fatal(err)
		}
	}

	return dbClient, dbConn
}
