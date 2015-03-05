package cropping

import (
	"file"
	"net/http"
	"strconv"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	"appengine"
	"appengine/datastore"
	af "appengine/file"

	"github.com/disintegration/imaging"
)

func init() {
	http.HandleFunc("/cropping", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	switch r.Method {
	default:
		http.Error(w, "unsupported method.", http.StatusMethodNotAllowed)
	case "GET":
		writeCroppingImage(c, w, r)
	}
}

func writeCroppingImage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var maxWidth int = 32
	var maxHeight int = 32

	id := r.FormValue("id")
	k := file.CreateBlobContentKey(c, id)

	pmw := r.FormValue("maxwidth")
	if pmw != "" {
		i64pmw, err := strconv.ParseInt(pmw, 10, 0)
		if err != nil {
			c.Errorf("invalid param maxwidth : %s", pmw)
			http.Error(w, "invalid param maxwidth", http.StatusBadRequest)
			return
		}
		maxWidth = int(i64pmw)
	}
	pmh := r.FormValue("maxheight")
	if pmh != "" {
		i64pmh, err := strconv.ParseInt(pmh, 10, 0)
		if err != nil {
			c.Errorf("invalid param maxheight : %s", pmh)
			http.Error(w, "invalid param maxheight", http.StatusBadRequest)
			return
		}
		maxHeight = int(i64pmh)
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

	var t *image.NRGBA
	if r.FormValue("cropcenter") != "" {
		t = imaging.CropCenter(img, maxWidth, maxHeight)
	} else {
		t = imaging.Thumbnail(img, maxWidth, maxHeight, imaging.Lanczos)
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	err = png.Encode(w, t)
	if err != nil {
		c.Errorf("failure thumnail image encode to http response : %s", err.Error())
		http.Error(w, "ailure thumnail image encode to http response", http.StatusInternalServerError)
		return
	}
}
