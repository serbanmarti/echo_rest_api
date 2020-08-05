package model

import (
	"context"

	"echo_rest_api/pkg/internal"
	"echo_rest_api/pkg/security"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	User struct {
		ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
		Email        string               `bson:"email" json:"email" validate:"required,email"`
		Password     string               `bson:"password" json:"password" validate:"required,password"`
		Salt         []byte               `bson:"salt" json:"-"`
		Role         string               `bson:"role" json:"role,omitempty" validate:"omitempty,role"`
		Active       bool                 `bson:"active" json:"-"`
		CreatedBy    primitive.ObjectID   `bson:"created_by" json:"-"`
		CreatedUsers []primitive.ObjectID `bson:"created_users" json:"-"`
		InviteToken  string               `bson:"invite_token,omitempty" json:"-"`
		Token        string               `bson:"-" json:"token,omitempty"`
	}

	UserMinimalData struct {
		ID     primitive.ObjectID `bson:"_id" json:"id"`
		Email  string             `bson:"email" json:"email"`
		Role   string             `bson:"role" json:"role"`
		Active bool               `bson:"active" json:"active"`
	}

	UserUpdateData struct {
		ID     primitive.ObjectID
		Role   string `json:"role" validate:"required"`
		Active bool   `json:"active" validate:"required"`
	}

	Invite struct {
		Email string `json:"email" validate:"required,email"`
		Role  string `json:"role" validate:"required,role"`
	}

	ValidateInvite struct {
		InviteToken string `json:"invite_token" validate:"required"`
	}

	SignUp struct {
		InviteToken string `json:"invite_token" validate:"required"`
		Password    string `json:"password" validate:"required,password"`
		Salt        []byte `json:"-"`
	}
)

const (
	usersCollectionName = "users"
)

// Check if a user is found based on given email and password in the DB
func UserFind(m *mongo.Database, u *User) error {
	// Save the raw password for later use
	rawPassword := u.Password

	// Create a DB connection
	db := m.Collection(usersCollectionName)

	if err := db.FindOne(context.TODO(), bson.M{"email": u.Email}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.NewDatabaseError(internal.ErrDBNoData, err, 1)
		}

		return internal.NewDatabaseError(internal.ErrDBQuery, err, 1)
	}

	// Create the hashed password for the current user
	hashedPassword := security.HashPassword(rawPassword, u.Salt)

	// Check if hashed passwords match
	if u.Password != hashedPassword {
		return internal.NewBackendError(internal.ErrBEInvalidPassword, nil, 1)
	}

	return nil
}

// Check if the root user account is already available in the DB
func UserFindRoot(m *mongo.Database, email string) (bool, error) {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	count, err := db.CountDocuments(context.TODO(), bson.M{"email": email})
	if err != nil && err != mongo.ErrNoDocuments {
		return false, internal.NewDatabaseError(internal.ErrDBQuery, err, 1)
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

// Check if a user account is already registered to a given email in the DB
func UserEmailExists(m *mongo.Database, i *Invite) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	count, err := db.CountDocuments(context.TODO(), bson.M{"email": i.Email})
	if err != nil && err != mongo.ErrNoDocuments {
		return internal.NewDatabaseError(internal.ErrDBQuery, err, 1)
	}

	if count > 0 {
		return internal.NewBackendError(internal.ErrBEUserExists, nil, 1)
	}

	return nil
}

// Create a new user in the DB
func UserCreate(m *mongo.Database, u *User) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	// Add the user to the DB
	newUser, err := db.InsertOne(context.TODO(), u)
	if err != nil {
		return internal.NewDatabaseError(internal.ErrDBInsert, err, 1)
	}

	// Save the ID of the new user
	u.ID = newUser.InsertedID.(primitive.ObjectID)

	// Update the creating user
	_, err = db.UpdateOne(context.TODO(), bson.M{"_id": u.CreatedBy}, bson.M{"$push": bson.M{"created_users": u.ID}})
	if err != nil {
		return internal.NewDatabaseError(internal.ErrDBUpdate, err, 1)
	}

	return nil
}

// Create a new root user in the DB
func UserCreateRoot(m *mongo.Database, u *User) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	// Add the user to the DB
	_, err := db.InsertOne(context.TODO(), u)
	if err != nil {
		return internal.NewDatabaseError(internal.ErrDBInsert, err, 1)
	}

	return nil
}

// Check if an invite token is still available in the DB
func UserValidateInvite(m *mongo.Database, i *ValidateInvite) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	count, err := db.CountDocuments(context.TODO(), bson.M{"invite_token": i.InviteToken})
	if err != nil && err != mongo.ErrNoDocuments {
		return internal.NewDatabaseError(internal.ErrDBQuery, err, 1)
	}

	if count == 0 {
		return internal.NewBackendError(internal.ErrBEInvalidInvite, nil, 1)
	}

	return nil
}

// Activate an invited account and set the password and salt in the DB
func UserSignUp(m *mongo.Database, s *SignUp) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	_, err := db.UpdateOne(context.TODO(), bson.M{"invite_token": s.InviteToken}, bson.M{
		"$set":   bson.M{"password": s.Password, "salt": s.Salt, "active": true},
		"$unset": bson.M{"invite_token": ""},
	})
	if err != nil {
		return internal.NewDatabaseError(internal.ErrDBUpdate, err, 1)
	}

	return nil
}

// Retrieve all users (except from the given ID) from the DB
func UserGetAll(m *mongo.Database, id primitive.ObjectID) ([]UserMinimalData, error) {
	var u []UserMinimalData

	// Create a DB connection
	db := m.Collection(usersCollectionName)

	// Find all users, except for the given ID
	cur, err := db.Find(context.TODO(), bson.M{"_id": bson.M{"$ne": id}})
	if err != nil {
		return nil, internal.NewDatabaseError(internal.ErrDBQuery, err, 1)
	}

	// Decode all found information
	for cur.Next(context.TODO()) {
		var elem UserMinimalData

		err = cur.Decode(&elem)
		if err != nil {
			return nil, internal.NewDatabaseError(internal.ErrDBDecode, err, 1)
		}

		u = append(u, elem)
	}

	// Check if any errors occurred
	if err = cur.Err(); err != nil {
		return nil, internal.NewDatabaseError(internal.ErrDBCursorIterate, err, 1)
	}

	// Close the cursor once finished
	if err = cur.Close(context.TODO()); err != nil {
		return nil, internal.NewDatabaseError(internal.ErrDBCursorClose, err, 1)
	}

	// Check if any data found
	if len(u) == 0 {
		return nil, internal.NewDatabaseError(internal.ErrDBNoData, err, 1)
	}

	return u, nil
}

// Update a given user data in the DB
func UserUpdate(m *mongo.Database, u *UserUpdateData) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	// Set the new values for active and role
	r, err := db.UpdateOne(context.TODO(), bson.M{"_id": u.ID}, bson.M{
		"$set": bson.M{"active": u.Active, "role": u.Role},
	})
	if err != nil {
		return internal.NewDatabaseError(internal.ErrDBUpdate, err, 1)
	}

	// If no data was updated, return an error
	if r.ModifiedCount == 0 {
		return internal.NewDatabaseError(internal.ErrDBNoUpdate, err, 1)
	}

	return nil
}

// Delete a user based on the given ID in the DB
func UserDelete(m *mongo.Database, id primitive.ObjectID) error {
	// Create a DB connection
	db := m.Collection(usersCollectionName)

	// Delete the user from the DB
	_, err := db.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		return internal.NewDatabaseError(internal.ErrDBDelete, err, 1)
	}

	return nil
}
