package handler

import (
	"fmt"
	"log"
	"net/http"
	"secnds-server/mailer"
	"secnds-server/model"
	"time"

	"net/mail"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

func checkStringAlphabet(str string) bool {
	for _, charVariable := range str {
		if (charVariable < 'a' || charVariable > 'z') && (charVariable < 'A' || charVariable > 'Z') {
			return false
		}
	}
	return true
}

func SignUp(userCollection *firestore.CollectionRef, userEmap *model.UsersEmailMap) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := new(model.User)
		if err := c.Bind(u); err != nil {
			return err
		}
		id := uuid.NewString()
		u.ID = id
		if u.EmailAddress == "" {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Email field cannot be empty"}

		}
		if u.Password == "" {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Password field cannot be empty"}
		}
		if u.FirstName == "" {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "First Name field cannot be empty"}

		}
		if u.LastName == "" {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Last Name field cannot be empty"}

		}
		_, err := mail.ParseAddress(u.EmailAddress)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Email address"}
		}
		usersEmailMap := *userEmap
		_, exists := usersEmailMap[u.EmailAddress]
		if exists {
			//If user already exists abort
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "User account with this email already exists !"}

		}
		if !checkStringAlphabet(u.FirstName) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "First Name contains invalid characters"}
		}
		if !checkStringAlphabet(u.LastName) {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Last Name contains invalid characters"}
		}
		formInputs := []string{u.EmailAddress, u.FirstName, u.LastName, u.MobileNumber}
		strength := zxcvbn.PasswordStrength(u.Password, formInputs)
		passwordScore := strength.Score
		fmt.Println(passwordScore)
		if passwordScore < 2 {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Password is weak"}

		}
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		if err != nil {
			log.Printf("Error hashing form password: %s", err)
		}
		u.Password = string(passwordHash)
		u.CreatedAt = time.Now()

		_, err = userCollection.Doc(u.ID).Set(c.Request().Context(), u)
		if err != nil {
			// Handle error when cannot store user data to Firestore DB
			log.Printf("Error occured when adding user to firestore DB : %s", err)
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error caused when storing data to DB"}
		}
		go mailer.SendMail(u.FirstName, u.EmailAddress, "Hello from secnds !")
		return c.JSON(http.StatusCreated, u)
	}

}
