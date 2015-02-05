package file

import (
	"appengine"
	"appengine/datastore"
	"appengine/file"

	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

type BlobContent struct {
	Id          string    `json: "id" datastore:"_"`
	Filename    string    `json: "filename"`
	AbsFilename string    `json: "absFilename"`
	ContentType string    `json: "contentType"`
	Size        int64     `json: "size"`
	CreatedAt   time.Time `json: "createdAt"`
}

func init() {
	http.HandleFunc("/file", handler)
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
	}
}

func downloadFile(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	bn, err := file.DefaultBucketName(c)
	if err != nil {
		c.Errorf("%s", err.Error())
	}

	// JSTの日ごとにPathを区切っておく
	filename := "/gs/" + bn + "/" + r.FormValue("name")
	log.Printf("filename : " + filename)
	fr, err := file.Open(c, filename)
	if err != nil {
		c.Errorf("%s", err.Error())
	}
	defer fr.Close()

	w.Header().Set("Cache-Control:public", "max-age=120")
	io.Copy(w, fr)
}

func directStore(c appengine.Context, f multipart.File, fh *multipart.FileHeader) (absFilename string, size int64, err error) {
	bn, err := file.DefaultBucketName(c)
	if err != nil {
		return "", 0, err
	}

	opts := &file.CreateOptions{
		MIMEType:   fh.Header.Get("Content-Type"),
		BucketName: bn,
	}

	wc, absFilename, err := file.Create(c, getNowDateJst(time.Now())+"/"+uuid.New(), opts)
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
