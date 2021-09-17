package model

import (
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	ID            string    `firestore:"id" json:"id"`
	EmailAddress  string    `firestore:"emailAddress" json:"emailAddress"`
	Password      string    `firestore:"password" json:"password"`
	MobileNumber  string    `firestore:"mobileNumber" json:"mobileNumber"`
	FirstName     string    `firestore:"firstName" json:"firstName"`
	LastName      string    `firestore:"lastName" json:"lastName"`
	postalAddress string    `firestore:"postalAddress" json:"postalAddress"`
	CreatedAt     time.Time `firestore:"createdAt" json:"createdAt"`
}

type Users []User
type UsersMap map[string]User
type UsersEmailMap map[string]User

func CreateUserInDB(userCollection *firestore.CollectionRef, c echo.Context, user *User) error {
	_, err := userCollection.Doc(user.ID).Set(c.Request().Context(), user)
	if err != nil {
		// Handle error when cannot store user data to Firestore DB
		log.Printf("Error occured when adding user to firestore DB : %s", err)
		return err
	}
	return nil
}

func FirestoreUserListen(user_snapshots *firestore.QuerySnapshotIterator, current_users_arr *Users, current_users_map *UsersMap, current_users_emap *UsersEmailMap) {
	for {
		snap, err := user_snapshots.Next()
		// DeadlineExceeded will be returned when ctx is cancelled.
		if status.Code(err) == codes.DeadlineExceeded {
			return
		}
		if err != nil {
			log.Fatalf("Snapshots.Next: %v", err)
		}
		if snap != nil {
			var new_users_arr Users
			var new_users_map UsersMap = make(UsersMap)
			var new_users_emap UsersEmailMap = make(UsersEmailMap)
			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					*current_users_arr = new_users_arr
					*current_users_map = new_users_map
					*current_users_emap = new_users_emap
					break
				}
				if err != nil {
					log.Fatalf("Documents.Next: %v", err)
				}
				//iter holds the listing of the current iteration
				var iter User
				if err := doc.DataTo(&iter); err != nil {
					log.Fatalf("Error occured when extracting data from firestore into Listing Struct: %v", err)
				}

				new_users_arr = append(new_users_arr, iter)
				var id string = iter.ID
				var email string = iter.EmailAddress
				new_users_map[id] = iter
				new_users_emap[email] = iter

			}
		}
	}
}
