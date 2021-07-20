package model

import (
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	Name string
}

type Listing struct {
	ID            string    `firestore:"id"`
	Title         string    `firestore:"title"`
	PostedAt      time.Time `firestore:"timestamp"`
	Price         string    `firestore:"price"`
	ListingImages []string  `firestore:"listingImages"`
}

type Listings []Listing

func (listings_arr *Listings) FirestoreListingListen(listing_snapshots *firestore.QuerySnapshotIterator) {
	for {
		snap, err := listing_snapshots.Next()
		// DeadlineExceeded will be returned when ctx is cancelled.
		if status.Code(err) == codes.DeadlineExceeded {
			return
		}
		if err != nil {
			log.Fatalf("Snapshots.Next: %v", err)
		}
		if snap != nil {
			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					log.Fatalf("Documents.Next: %v", err)
				}
				//iter holds the listing of the current iteration
				var iter Listing
				if err := doc.DataTo(&iter); err != nil {
					log.Fatalf("Error occured when extracting data from firestore into Listing Struct: %v", err)
				}

				*listings_arr = append(*listings_arr, iter)

			}
		}
	}
}
