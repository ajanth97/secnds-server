package handler

import (
	"log"
	"net/http"
	"secnds-server/mailer"
	"secnds-server/model"
	"secnds-server/token"
	"time"

	"net/mail"

	"cloud.google.com/go/firestore"
	jwtv3 "github.com/dgrijalva/jwt-go"
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

func emailExists(email string, userEmap *model.UsersEmailMap) bool {
	usersEmailMap := *userEmap
	_, exists := usersEmailMap[email]
	return exists
}

func passwordValid(email string, password string, userEmap *model.UsersEmailMap) bool {
	usersEmailMap := *userEmap
	actualUser, _ := usersEmailMap[email]
	actualPassword := actualUser.Password
	if (bcrypt.CompareHashAndPassword([]byte(actualPassword), []byte(password))) == nil {
		return true
	}
	return false
}

func getUserFromEmail(email string, userEmap *model.UsersEmailMap) model.User {
	usersEmailMap := *userEmap
	user, exists := usersEmailMap[email]
	if !exists {
		log.Fatalf("User doesnt exist in User Email map")
	}
	return user
}

func getUserFromId(id string, userMap *model.UsersMap) model.User {
	usersIdMap := *userMap
	user, exists := usersIdMap[id]
	if !exists {
		log.Fatalf("User doesnt exist in User ID map")
	}
	return user
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
		if emailExists(u.EmailAddress, userEmap) {
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
		jwtToken := token.GetJwtToken(u.ID, u.EmailAddress)
		go mailer.SendMail(u.FirstName, u.EmailAddress, "Hello from secnds !")
		return c.JSON(http.StatusCreated, jwtToken)
	}
}

func Login(userEmap *model.UsersEmailMap) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")
		if emailExists(email, userEmap) && passwordValid(email, password, userEmap) {
			user := getUserFromEmail(email, userEmap)
			jwtToken := token.GetJwtToken(user.ID, user.EmailAddress)
			return c.JSON(http.StatusOK, echo.Map{
				"token": jwtToken,
			})
		}
		return c.JSON(http.StatusUnauthorized, "Invalid Email or Password")
	}
}

func MyAccount(userMap *model.UsersMap) echo.HandlerFunc {
	return func(c echo.Context) error {
		userToken := c.Get("user").(*jwtv3.Token)
		//type asserting to string
		userId := userToken.Claims.(jwtv3.MapClaims)["id"].(string)
		user := getUserFromId(userId, userMap)
		return c.JSON(http.StatusOK, user)
	}
}
