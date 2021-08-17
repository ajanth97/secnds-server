package model

import (
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Listing struct {
	ID            string    `firestore:"id"`
	Title         string    `firestore:"title"`
	PostedAt      time.Time `firestore:"timestamp"`
	Price         string    `firestore:"price"`
	ListingImages []string  `firestore:"listingImages"`
}

type Listings []Listing
type ListingsMap map[string]Listing

func FirestoreListingListen(listing_snapshots *firestore.QuerySnapshotIterator, current_listings_arr *Listings, current_listings_map *ListingsMap) {
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
			var new_listings_arr Listings
			var new_listings_map ListingsMap = make(ListingsMap)
			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					*current_listings_arr = new_listings_arr
					*current_listings_map = new_listings_map
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

				new_listings_arr = append(new_listings_arr, iter)
				var id string = iter.ID
				new_listings_map[id] = iter

			}
		}
	}
}
