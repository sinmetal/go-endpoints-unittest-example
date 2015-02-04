package sample

import (
	"net/http"
	"net/http/httptest"
	"time"

	"testing"

	"appengine"
	"appengine/user"
	"appengine/aetest"
	"appengine/datastore"

	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
)

type unitTestContext struct {
	appengine.Context
}

func (c *unitTestContext) HTTPRequest() *http.Request {
	return c.Request().(*http.Request)
}

func (c *unitTestContext) Namespace(name string) (endpoints.Context, error) {
	nc, err := appengine.Namespace(c, name)
	if err != nil {
		return nil, err
	}
	return &unitTestContext{nc}, nil
}

func (c *unitTestContext) CurrentOAuthClientID(scope string) (string, error) {
	return "", nil
}

func (c *unitTestContext) SetContext(context appengine.Context) {
	c.Context = context
}

func (c *unitTestContext) CurrentOAuthUser(scope string) (*user.User, error) {
	return nil, nil
}

func stubContextFactory(r *http.Request) endpoints.Context {
	t := &unitTestContext{}
	t.Context = testContext
	return t
}

var testContext appengine.Context

func TestSimpleDatastoreOpe(t *testing.T) {
	opt := aetest.Options{AppID:"unittest", StronglyConsistentDatastore: true}
	c, err := aetest.NewContext(&opt)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Close()
	testContext = c

	origFactory := endpoints.ContextFactory
	endpoints.ContextFactory = stubContextFactory
	defer func() {
		endpoints.ContextFactory = origFactory
	}()

	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer instance.Close()

	key := datastore.NewKey(c, "Greeting", "testid", 0, nil)
	g := &Greeting{
		key,
		"sinmetal",
		"Hello! go endpoints.",
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	}
	_, err = datastore.Put(c, key, g)
	if err != nil {
		t.Fatal(err.Error())
	}

	stored := &Greeting{}
	err = datastore.Get(c, key, stored)
	if err != nil {
		t.Fatal(err.Error())
	}

	const limit int = 10
	q := datastore.NewQuery("Greeting").Limit(limit)
	greets := make([]*Greeting, 0, limit)
	keys, err := q.GetAll(c, &greets)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(keys) != 1 {
		t.Fatalf("Non-expected Query Results GreetingList lenght : %v", len(keys))
	}
}

func TestList(t *testing.T) {
	opt := aetest.Options{AppID:"unittest", StronglyConsistentDatastore: true}
	c, err := aetest.NewContext(&opt)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Close()
	testContext = c

	origFactory := endpoints.ContextFactory
	endpoints.ContextFactory = stubContextFactory
	defer func() {
		endpoints.ContextFactory = origFactory
	}()

	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer instance.Close()

	key := datastore.NewKey(c, "Greeting", "testid", 0, nil)
	g := &Greeting{
		key,
		"sinmetal",
		"Hello! go endpoints.",
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	}
	_, err = datastore.Put(c, key, g)
	if err != nil {
		t.Fatal(err.Error())
	}

	// try get
	stored := &Greeting{}
	err = datastore.Get(c, key, stored)
	if err != nil {
		t.Fatal(err.Error())
	}

	// try query
	const limit int = 10
	q := datastore.NewQuery("Greeting").Limit(limit)
	greets := make([]*Greeting, 0, limit)
	keys, err := q.GetAll(c, &greets)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(keys) != 1 {
		t.Fatalf("Non-expected Query Results GreetingList lenght : %v", len(keys))
	}

	req, err := instance.NewRequest("GET", "/_ah/api/greeting/v1/greetings/", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	res := httptest.NewRecorder()

	ec := endpoints.NewContext(req)

	glr := &GreetingsListReq{10}

	gs := GreetingService{}
	gl, err := gs.List(ec, glr)
	if err != nil {
		t.Fatal(err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Non-expected status code : %v\n\tbody: %v", res.Code, res.Body)
	}

	if len(gl.Items) != 1 {
		t.Fatalf("Non-expected Response GreetingList lenght : %v", len(gl.Items))
	}
}
