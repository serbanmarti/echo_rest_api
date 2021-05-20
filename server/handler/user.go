package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/gomail.v2"

	"echo_rest_api/database/model"
	"echo_rest_api/internal"
	"echo_rest_api/security"
)

// GET

// UserGetAll gets all available users data (except the logged in one)
func (h *Handler) UserGetAll(c echo.Context) (err error) {
	// Get authenticated user data
	claims := internal.DecodeClaims(c)

	// Get the logged-in user ID
	id, err := claims.GetUserID()
	if err != nil {
		return
	}

	// Retrieve all users data from the DB (except the logged in one)
	u, err := model.UserGetAll(h.DB, id)
	if err != nil {
		return
	}

	return HTTPSuccess(c, u)
}

// POST

// Login a user into the system and return an authorization JWT
func (h *Handler) Login(c echo.Context) (err error) {
	// Bind request data
	u := new(model.User)
	if err = c.Bind(u); err != nil {
		return
	}

	// Validate request data
	if err = c.Validate(u); err != nil {
		return
	}

	// Find the user in the DB
	if err = model.UserFind(h.DB, u); err != nil {
		return
	}

	// Check if the account is active
	if !u.Active {
		return internal.NewError(internal.ErrBENotActive, nil, 1)
	}

	// Remove the password from memory (this is returned to the requester)
	u.Password = ""

	// Get input query parameters
	qp := c.QueryParams()

	// Parse mobile request
	m := false
	mRaw := qp.Get("mobile")
	if mRaw != "" {
		m, err = strconv.ParseBool(mRaw)
		if err != nil {
			return internal.NewError(internal.ErrBEQPInvalidMobile, nil, 1)
		}
	}

	// JWT

	// If we have a mobile request, set a long expiration time
	var exp time.Duration
	if exp = h.JwtExp; m {
		exp, _ = time.ParseDuration("87600h")
	}

	// Set claims
	claims := &internal.JWTClaims{
		Role: u.Role,
		StandardClaims: jwt.StandardClaims{
			Id:        u.ID.Hex(),
			ExpiresAt: time.Now().Add(exp).Unix(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Generate encoded token and send it as response.
	u.Token, err = token.SignedString([]byte(h.JwtSecret))
	if err != nil {
		return
	}

	return HTTPSuccess(c, u)
}

// Invite a new user into the system
func (h *Handler) Invite(c echo.Context) (err error) {
	// Get authenticated user data
	claims := internal.DecodeClaims(c)

	// Check if the user has rights to this section
	if err = claims.IsAdmin(); err != nil {
		return
	}

	// Bind request data
	i := new(model.Invite)
	if err = c.Bind(i); err != nil {
		return
	}

	// Validate request data
	if err = c.Validate(i); err != nil {
		return
	}

	// Check if the given email not already registered
	if err = model.UserEmailExists(h.DB, i); err != nil {
		return
	}

	// Create the new user object
	u := &model.User{
		Email:        i.Email,
		Role:         i.Role,
		Active:       false,
		CreatedUsers: []primitive.ObjectID{},
	}

	// Create an invite token for the new user
	u.InviteToken = uuid.New().String()

	// Get the logged-in user ID to set them as the creator
	u.CreatedBy, err = claims.GetUserID()
	if err != nil {
		return
	}

	// Create a new user and add them to the DB
	if err = model.UserCreate(h.DB, u); err != nil {
		return
	}

	// Send the invite email
	if h.SMTP.Host != "" {
		// Instantiate a new message
		m := gomail.NewMessage()

		// Set the required headers
		m.SetHeaders(map[string][]string{
			"From":    {m.FormatAddress(h.SMTP.User, "REST.API")},
			"To":      {i.Email},
			"Subject": {"Invited to REST.API!"},
		})

		// Set the body of the email
		b := fmt.Sprintf(
			"Hello!<br><br>Click on the following link to activate your account: %s/register/:?token=%s.",
			h.FEndpoint, u.InviteToken,
		)
		m.SetBody("text/html", b)

		d := gomail.NewDialer(h.SMTP.Host, h.SMTP.Port, h.SMTP.User, h.SMTP.Pass)
		if err := d.DialAndSend(m); err != nil {
			return internal.NewError(internal.ErrBEEmail, err, 1)
		}
	}

	return HTTPSuccess(c, map[string]interface{}{
		"id": u.ID,
	})
}

// ValidateInvite validates a user invite token
func (h *Handler) ValidateInvite(c echo.Context) (err error) {
	// Bind request data
	i := new(model.ValidateInvite)
	if err = c.Bind(i); err != nil {
		return
	}

	// Validate request data
	if err = c.Validate(i); err != nil {
		return
	}

	// Check if the given invite token is still available
	if err = model.UserValidateInvite(h.DB, i); err != nil {
		return
	}

	return HTTPSuccess(c, nil)
}

// SignUp activates an invited user account
func (h *Handler) SignUp(c echo.Context) (err error) {
	// Bind request data
	s := new(model.SignUp)
	if err = c.Bind(s); err != nil {
		return
	}

	// Validate request data
	if err = c.Validate(s); err != nil {
		return
	}

	// Generate a new random salt
	s.Salt, err = security.NewSalt()
	if err != nil {
		return
	}

	// Hash and set the password field
	s.Password = security.HashPassword(s.Password, s.Salt)

	// Activate the invited user in the DB
	if err = model.UserSignUp(h.DB, s); err != nil {
		return
	}

	return HTTPSuccess(c, nil)
}

// PUT

// UserUpdate updates user data for a given ID
func (h *Handler) UserUpdate(c echo.Context) (err error) {
	// Get authenticated user data
	claims := internal.DecodeClaims(c)

	// Check if the user has rights to this section
	if err = claims.IsAdmin(); err != nil {
		return
	}

	// Bind request data
	u := new(model.UserUpdateData)

	u.ID, err = internal.DecodeParameterID(c, "userID")
	if err != nil {
		return
	}

	if err = c.Bind(u); err != nil {
		return
	}

	// Validate request data
	if err = c.Validate(u); err != nil {
		return
	}

	// Update the user details
	if err = model.UserUpdate(h.DB, u); err != nil {
		return
	}

	return HTTPSuccess(c, nil)
}

// DELETE

// UserDelete deletes a user from a given ID
func (h *Handler) UserDelete(c echo.Context) (err error) {
	// Get authenticated user data
	claims := internal.DecodeClaims(c)

	// Check if the user has rights to this section
	if err = claims.IsAdmin(); err != nil {
		return
	}

	// Bind request data
	id, err := internal.DecodeParameterID(c, "userID")
	if err != nil {
		return
	}

	// Delete the user in the DB
	if err = model.UserDelete(h.DB, id); err != nil {
		return
	}

	return HTTPSuccess(c, nil)
}
