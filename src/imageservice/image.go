package imageservice

import (
	"encoding/json"
	"net/http"

	"file"

	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"appengine/image"
)

func init() {
	http.HandleFunc("/image", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	switch r.Method {
	default:
		http.Error(w, "unsupported method.", http.StatusMethodNotAllowed)
	case "GET":
		getImageUrl(c, w, r)
	}
}

func getImageUrl(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	k := file.CreateBlobContentKey(c, id)

	var b file.BlobContent
	err := datastore.Get(c, k, &b)
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bk, err := blobstore.BlobKeyForFile(c, b.AbsFilename)
	if err != nil {
		c.Errorf("failure, create BlobKey, %s", b.AbsFilename)
		http.Error(w, "failure create BlobKey", http.StatusInternalServerError)
		return
	}
	url, err := image.ServingURL(c, bk, nil)
	if err != nil {
		c.Errorf("failure, image serving url, %s", bk)
		http.Error(w, "failure image serving url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(url)
	if err != nil {
		c.Errorf("failure, encode url to json, %s", err.Error())
		http.Error(w, "failure encode url to json", http.StatusInternalServerError)
		return
	}
}
