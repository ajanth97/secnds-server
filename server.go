package main

import (
	"context"

	"secnds-server/env"
	"secnds-server/handler"
	"secnds-server/model"

	firebase "firebase.google.com/go/v4"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
)

const (
	jwt_secret      = "JWT_SECRET"
	port_           = "PORT"
	google_key      = "GOOGLE_KEY"
	google_secret   = "GOOGLE_SECRET"
	facebook_key    = "FACEBOOK_KEY"
	facebook_secret = "FACEBOOK_SECRET"
)

var jwtSecret []byte = env.GetByte(jwt_secret)
var port string = env.Get(port_)
var googleKey = env.Get(google_key)
var googleSecret = env.Get(google_secret)
var facebookKey = env.Get(facebook_key)
var facebookSecret = env.Get(facebook_secret)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://127.0.0.1:8000", "https://alpha.secnds.com"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlRequestHeaders, echo.HeaderAccessControlAllowOrigin, echo.HeaderAuthorization},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST"},
	}))

	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: jwtSecret,
		Skipper: func(c echo.Context) bool {
			// Skip authentication for signup and login requests
			if c.Path() == "/login" || c.Path() == "/signup" || c.Path() == "/listing/all" || c.Path() == "/listing/add" || c.Path() == "/listing/:id" || c.Path() == "/auth/:provider" || c.Path() == "/auth/:provider/callback" {
				return true
			}
			return false
		},
	}))

	//e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	store := sessions.NewCookieStore([]byte("secret"))
	gothic.Store = store
	//	e.Use(session.Middleware(store))
	// Database connection
	ctx := context.Background()
	//Initializing Firebase
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	fs, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error initializing database client:", err)
	}

	user_collections := fs.Collection("users")
	user_snapshots := user_collections.Snapshots(ctx)
	var users model.Users
	var usersMap model.UsersMap
	var usersEmailMap model.UsersEmailMap
	go model.FirestoreUserListen(user_snapshots, &users, &usersMap, &usersEmailMap)

	listing_snapshots := fs.Collection("listings").Snapshots(ctx)
	var listings model.Listings
	var listingsMap model.ListingsMap
	go model.FirestoreListingListen(listing_snapshots, &listings, &listingsMap)

	goth.UseProviders(google.New(googleKey, googleSecret, "https://secnds-server.herokuapp.com/auth/google/callback", "email", "https://www.googleapis.com/auth/userinfo.profile"))
	goth.UseProviders(facebook.New(facebookKey, facebookSecret, "https://secnds-server.herokuapp.com/auth/facebook/callback"))

	// Routes
	e.POST("/signup", handler.SignUp(user_collections, &usersEmailMap))
	e.POST("/login", handler.Login(&usersEmailMap))
	e.GET("/auth/:provider", handler.LoginWithThirdParty())
	e.GET("/auth/:provider/callback", handler.LoginWithThirdPartyCallBack(user_collections, &usersEmailMap))
	//e.POST("/follow/:id", h.Follow)
	//e.POST("/posts", h.CreatePost)
	//e.GET("/feed", h.FetchPost)
	e.GET("/listing/:id", handler.FetchListing(&listingsMap))
	e.GET("/listing/all", handler.FetchAllListings(&listings))
	e.POST("/listing/add", handler.CreateListing)
	e.GET("/myaccount", handler.MyAccount(&usersMap))

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}
