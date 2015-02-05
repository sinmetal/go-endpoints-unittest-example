package file

import (
	"appengine"
	"appengine/file"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

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
			fmt.Fprint(w, "Upload the file is too large.\n")
		} else {
			c.Errorf("%s", err.Error())
		}
		return
	}
	f, fh, err := r.FormFile("filename")
	if err != nil {
		c.Errorf("%s", err.Error())
		return
	}
	defer f.Close()

	log.Printf("%s", fh.Filename)
	log.Printf("%v", fh.Header)
	log.Printf("Content-Type : %s", fh.Header.Get("Content-Type"))

	absFilename, size, err := directStore(c, f, fh)
	if err != nil {
		c.Errorf("%s", err.Error())
		return
	}

	log.Printf("absFilename : %s", absFilename)
	log.Printf("size : %d", size)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(absFilename))
}

func downloadFile(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	bn, err := file.DefaultBucketName(c)
	if err != nil {
		c.Errorf("%s", err.Error())
	}

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

	wc, absFilename, err := file.Create(c, fh.Filename, opts)
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
