package main

import (
	"context"

	"secnds-server/handler"
	"secnds-server/model"

	firebase "firebase.google.com/go/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	key := "key"
	e := echo.New()
	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Logger())
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(key),
		Skipper: func(c echo.Context) bool {
			// Skip authentication for signup and login requests
			if c.Path() == "/login" || c.Path() == "/signup" || c.Path() == "/listing/all" || c.Path() == "/listing/add" || c.Path() == "/listing/:id" {
				return true
			}
			return false
		},
	}))

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

	// Routes
	e.POST("/signup", handler.SignUp(user_collections, &usersEmailMap))
	//e.POST("/login", h.Login)
	//e.POST("/follow/:id", h.Follow)
	//e.POST("/posts", h.CreatePost)
	//e.GET("/feed", h.FetchPost)
	e.GET("/listing/:id", handler.FetchListing(&listingsMap))
	e.GET("/listing/all", handler.FetchAllListings(&listings))
	e.POST("/listing/add", handler.CreateListing)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
