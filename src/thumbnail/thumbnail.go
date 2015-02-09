package thumbnail

import (
	"file"
	"net/http"
	"strconv"

	"appengine"
	"appengine/datastore"
	af "appengine/file"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
)

func init() {
	http.HandleFunc("/thumbnail", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	switch r.Method {
	default:
		http.Error(w, "unsupported method.", http.StatusMethodNotAllowed)
	case "GET":
		writeThumbnail(c, w, r)
	}
}

func writeThumbnail(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var maxWidth uint = 32
	var maxHeight uint = 32

	id := r.FormValue("id")
	k := file.CreateBlobContentKey(c, id)

	pmw := r.FormValue("maxwidth")
	if pmw != "" {
		u64pmw, err := strconv.ParseUint(pmw, 10, 0)
		if err != nil {
			c.Errorf("invalid param maxwidth : %s", pmw)
			http.Error(w, "invalid param maxwidth", http.StatusBadRequest)
			return
		}
		maxWidth = uint(u64pmw)
	}
	pmh := r.FormValue("maxheight")
	if pmh != "" {
		u64pmh, err := strconv.ParseUint(pmh, 10, 0)
		if err != nil {
			c.Errorf("invalid param maxheight : %s", pmh)
			http.Error(w, "invalid param maxheight", http.StatusBadRequest)
			return
		}
		maxHeight = uint(u64pmh)
	}

	var b file.BlobContent
	err := datastore.Get(c, k, &b)
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fr, err := af.Open(c, b.AbsFilename)
	if err != nil {
		c.Errorf("failure open file : %s", b.AbsFilename)
		http.Error(w, "failure open file", http.StatusInternalServerError)
		return
	}

	img, _, err := image.Decode(fr)
	if err != nil {
		c.Errorf("failure decode image file : %s", b.AbsFilename)
		http.Error(w, "failure decode image file", http.StatusInternalServerError)
		return
	}

	t := resize.Thumbnail(maxWidth, maxHeight, img, resize.Lanczos3)

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	err = png.Encode(w, t)
	if err != nil {
		c.Errorf("failure thumnail image encode to http response : %s", err.Error())
		http.Error(w, "ailure thumnail image encode to http response", http.StatusInternalServerError)
		return
	}
}
