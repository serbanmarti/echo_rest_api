package security

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"echo_rest_api/internal"
)

func ConfigureEchoSecurity(e *echo.Echo, env *internal.Environ) {
	// Configure Echo JWT authorization
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:        &internal.JWTClaims{},
		SigningKey:    []byte(env.JwtSecret),
		SigningMethod: "HS512",
		Skipper: func(c echo.Context) bool {
			// Skip authentication for certain request routes
			return internal.InSlice(c.Path(), []string{"/", "/login", "/validate_invite", "/signup"})
		},
	}))

	// Get the required cookie domain
	cookieDomain, cErr := internal.GetDomainFromURL(env.FEndpoint)
	if cErr != nil {
		e.Logger.Fatal(cErr)
	}

	// Configure Echo CSRF checking
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:  "header:X-CSRF-Token",
		CookiePath:   "/",
		CookieDomain: cookieDomain,
		CookieSecure: env.Secure,
		Skipper: func(c echo.Context) bool {
			// Skip CSRF for certain request routes
			return internal.InSlice(c.Path(), []string{"/login"})
		},
	}))

	// Configure Echo extra security
	e.Use(middleware.Secure())
}
