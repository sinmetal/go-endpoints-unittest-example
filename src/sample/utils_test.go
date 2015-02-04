package sample

import (
	"net/http"

	"appengine"
	"appengine/user"

	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
)

type unitTestContext struct {
	appengine.Context
}

var testContext appengine.Context

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
