package sample

import (
	"appengine/datastore"
	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
	sessions "github.com/hnakamur/gaesessions"
	"time"
	"net/http"
)

const sessionName = "testcookie"

var memcacheDatastoreStore = sessions.NewMemcacheDatastoreStore("", "", sessions.DefaultNonPersistentSessionDuration, []byte("hogehogefugafuga"))

// Greeting is a datastore entity that represents a single greeting.
// It also serves as (a part of) a response of GreetingService.
type Greeting struct {
	Key     *datastore.Key `json:"id" datastore:"-"`
	Author  string         `json:"author"`
	Content string         `json:"content" datastore:",noindex" endpoints:"req"`
	Date    time.Time      `json:"date"`
}

// GreetingsList is a response type of GreetingService.List method
type GreetingsList struct {
	Items []*Greeting `json:"items"`
}

// Request type for GreetingService.List
type GreetingsListReq struct {
	Limit int `json:"limit" endpoints:"d=10"`
}

// GreetingService can sign the guesbook, list all greetings and delete
// a greeting from the guestbook.
type GreetingService struct {
}

func init() {
	http.HandleFunc("/cookie", saveCookie)

	greetService := &GreetingService{}
	api, err := endpoints.RegisterService(greetService,
		"greeting", "v1", "Greetings API", true)
	if err != nil {
		panic(err.Error())
	}

	info := api.MethodByName("List").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"greets.list", "GET", "greetings", "List most recent greetings."

	endpoints.HandleHTTP()
}

// List responds with a list of all greetings ordered by Date field.
// Most recent greets come first.
func (gs *GreetingService) List(c endpoints.Context, r *GreetingsListReq) (*GreetingsList, error) {
	if r.Limit <= 0 {
		r.Limit = 10
	}

	q := datastore.NewQuery("Greeting").Limit(r.Limit)
	greets := make([]*Greeting, 0, r.Limit)
	keys, err := q.GetAll(c, &greets)
	if err != nil {
		return nil, err
	}

	for i, k := range keys {
		greets[i].Key = k
	}
	return &GreetingsList{greets}, nil
}

func saveCookie(w http.ResponseWriter, r *http.Request) {
	session, err := memcacheDatastoreStore.Get(r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["id"] = "testcookievalue"
	memcacheDatastoreStore.Save(r, w, session)
}
