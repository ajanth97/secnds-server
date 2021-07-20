package handler

import (
	"net/http"
	"secnds-server/model"
	"time"

	"github.com/labstack/echo/v4"
)

func FetchAllListings(l *model.Listings) echo.HandlerFunc {

	return func(c echo.Context) error {
		// Retrieve posts from Listings pointer
		return c.JSON(http.StatusOK, l)
	}

}

func CreateListing(c echo.Context) error {
	l := new(model.Listing)
	if err := c.Bind(l); err != nil {
		return err
	}

	//Validation
	if l.Title == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Title cannot be blank"}
	}
	if l.Price == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Price cannot be blank"}
	}

	l.PostedAt = time.Now()
	return c.JSON(http.StatusCreated, l)
}
