package file

import (
	"appengine"
	"appengine/datastore"
	af "appengine/file"

	"file"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)

type BlobContent struct {
	Id          string    `json: "id" datastore:"_"`
	Filename    string    `json: "filename" datastore:"noindex"`
	AbsFilename string    `json: "absFilename" datastore:"noindex"`
	ContentType string    `json: "contentType" datastore:"noindex"`
	Size        int64     `json: "size" datastore:"noindex"`
	CreatedAt   time.Time `json: "createdAt"`
}

func init() {
	http.HandleFunc("/file", handler)
	http.HandleFunc("/fileMulti", handlerMulti)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	switch r.Method {
	default:
		http.Error(w, "unsupported method.", http.StatusMethodNotAllowed)
	case "POST":
		uploadFile(c, w, r)
	case "GET":
		downloadFile(c, w, r)
	}
}

func handlerMulti(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	switch r.Method {
	default:
		http.Error(w, "unsupported method.", http.StatusMethodNotAllowed)
	case "POST":
		uploadFileMulti(c, w, r)
	}
}

func uploadFile(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1024 * 1024) // 10MB
	if err != nil {
		if err.Error() == "permission denied" {
			http.Error(w, "Upload the file is too large.\n", http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	f, fh, err := r.FormFile("filename")
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	log.Printf("%s", fh.Filename)
	log.Printf("%v", fh.Header)
	log.Printf("Content-Type : %s", fh.Header.Get("Content-Type"))

	absFilename, size, err := directStore(c, f, fh)
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("absFilename : %s", absFilename)
	log.Printf("size : %d", size)

	b := &BlobContent{
		uuid.New(),
		fh.Filename,
		absFilename,
		fh.Header.Get("Content-Type"),
		size,
		time.Now(),
	}
	_, err = b.Save(c)
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func uploadFileMulti(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1024 * 1024) // grab the multipart form
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	formdata := r.MultipartForm  // ok, no problem so far, read the Form data

	//get the *fileheaders
	files := formdata.File["multiplefiles"]  // grab the filenames

	for i, _ := range files {  // loop through the files one by one
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		absFilename, size, err := directStore(c, file, files[i])
		if err != nil {
			c.Errorf("%s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}


		fmt.Fprintf(w,"%s, size = %d", absFilename, size)
	}

	w.WriteHeader(http.StatusOK)
}

func downloadFile(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	k := CreateBlobContentKey(c, id)

	var b BlobContent
	err := datastore.Get(c, k, &b)
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ims := r.Header.Get("If-Modified-Since")
	if ims != "" {
		imsTime, err := parseTime(ims)
		if err != nil {
			c.Errorf("If-Modified-Since Parse Error : %v \n %s", ims, err.Error())
		} else {
			if b.CreatedAt.Equal(imsTime) || b.CreatedAt.After(imsTime) {
				w.Header().Set("Last-Modified", b.CreatedAt.String())
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	fr, err := af.Open(c, b.AbsFilename)
	if err != nil {
		c.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fr.Close()

	w.Header().Set("Cache-Control:public", "max-age=120")
	w.Header().Set("Content-Type", b.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(b.Size, 10))
	w.Header().Set("Last-Modified", b.CreatedAt.String())
	io.Copy(w, fr)
}

func directStore(c appengine.Context, f multipart.File, fh *multipart.FileHeader) (absFilename string, size int64, err error) {
	bn, err := af.DefaultBucketName(c)
	if err != nil {
		return "", 0, err
	}

	opts := &af.CreateOptions{
		MIMEType:   fh.Header.Get("Content-Type"),
		BucketName: bn,
	}

	// JSTで、日ごとにPathを区切っておく
	wc, absFilename, err := af.Create(c, getNowDateJst(time.Now())+"/"+uuid.New(), opts)
	if err != nil {
		return "", 0, err
	}
	defer wc.Close()

	size, err = io.Copy(wc, f)
	if err != nil {
		return "", 0, err
	}

	return absFilename, size, nil
}

func getNowDateJst(t time.Time) string {
	j := t.In(time.FixedZone("Asia/Tokyo", 9*60*60))
	return j.Format("20060102")
}

const timeFormat = "2006-01-02 15:04:05.99999 -0700 MST"

var timeFormats = []string{
	time.RFC1123,
	time.RFC1123Z,
	timeFormat,
	time.RFC850,
	time.ANSIC,
}

func parseTime(text string) (t time.Time, err error) {
	for _, layout := range timeFormats {
		t, err = time.Parse(layout, text)
		if err == nil {
			return
		}
	}
	return
}

func CreateBlobContentKey(c appengine.Context, id string) *datastore.Key {
	log.Printf("key name = %s : ", id)
	return datastore.NewKey(c, "BlobContent", id, 0, nil)
}

func (b *BlobContent) Key(c appengine.Context) *datastore.Key {
	return CreateBlobContentKey(c, b.Id)
}

func (b *BlobContent) Save(c appengine.Context) (*BlobContent, error) {
	b.CreatedAt = time.Now()
	k, err := datastore.Put(c, b.Key(c), b)
	if err != nil {
		return nil, err
	}

	b.Id = k.StringID()
	return b, nil
}
