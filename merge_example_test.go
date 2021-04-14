package pipeline_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/db"
)

// SearchResults returns many types of search results at once
type SearchResults struct {
	Advertisements []db.Result `json:"advertisements"`
	Images         []db.Result `json:"images"`
	Products       []db.Result `json:"products"`
	Websites       []db.Result `json:"websites"`
}

func ExampleMerge() {
	r := http.NewServeMux()

	// `GET /search?q=<query>` is an endpoint that merges concurrently fetched
	// search results into a single search response using `pipeline.Merge`
	r.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if len(query) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// If the request times out, or we receive an error from our `db`
		// the context will stop all pending db queries for this request
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Fetch all of the different search results concurrently
		var results SearchResults
		for err := range pipeline.Merge(
			db.GetAdvertisements(ctx, query, &results.Advertisements),
			db.GetImages(ctx, query, &results.Images),
			db.GetProducts(ctx, query, &results.Products),
			db.GetWebsites(ctx, query, &results.Websites),
		) {
			// Stop all pending db requests if theres an error
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// Return the search results
		if bs, err := json.Marshal(&results); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if _, err := w.Write(bs); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}
