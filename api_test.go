package humanitec

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/humanitec/humanitec-go-autogen/client"
	"github.com/stretchr/testify/assert"
)

func TestNewHumanitecClientRead(t *testing.T) {
	assert := assert.New(t)

	expected := "{}"
	token := "TEST_TOKEN"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(fmt.Sprintf("Bearer %s", token), r.Header.Get("Authorization"))
		assert.Equal("sdk humanitec-go-autogen/latest", r.Header.Get("Humanitec-User-Agent"))
		fmt.Fprint(w, expected)
	}))
	defer srv.Close()

	ctx := context.Background()

	humSvc, err := NewClient(&Config{
		Token: token,
		URL:   srv.URL,
	})
	assert.NoError(err)

	_, err = humSvc.GetCurrentUser(ctx)
	assert.NoError(err)
}

func TestNewHumanitecClientWrite(t *testing.T) {
	assert := assert.New(t)

	expected := "{}"
	token := "TEST_TOKEN"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(fmt.Sprintf("Bearer %s", token), r.Header.Get("Authorization"))
		assert.Equal("app test/latest; sdk humanitec-go-autogen/latest", r.Header.Get("Humanitec-User-Agent"))

		defer r.Body.Close()
		resBody, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		assert.Equal("{\"name\":\"changed\"}", string(resBody))

		fmt.Fprint(w, expected)
	}))
	defer srv.Close()

	ctx := context.Background()

	humSvc, err := NewClient(&Config{
		Token:       token,
		URL:         srv.URL,
		InternalApp: "test/latest",
	})
	assert.NoError(err)

	name := "changed"
	_, err = humSvc.UpdateCurrentUser(ctx, client.UpdateCurrentUserJSONRequestBody{
		Name: &name,
	})
	assert.NoError(err)
}

func TestClient(t *testing.T) {
	assert := assert.New(t)

	token := "TEST_TOKEN"
	url := "https://my-test/"

	humSvc, err := NewClient(&Config{
		Token:       token,
		URL:         url,
		InternalApp: "test/latest",
	})
	assert.NoError(err)

	assert.Equal(url, humSvc.Client().Server)
}

func TestNewHumanitecClientMissingToken(t *testing.T) {
	assert := assert.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail("Shouldn't be called")
	}))
	defer srv.Close()

	_, err := NewClient(&Config{
		URL: srv.URL,
	})
	assert.ErrorIs(err, ErrMissingToken)
}

func TestNewHumanitecClientMissingTokenSkipCheck(t *testing.T) {
	assert := assert.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail("Shouldn't be called")
	}))
	defer srv.Close()

	_, err := NewClient(&Config{
		URL:                   srv.URL,
		SkipInitialTokenCheck: true,
	})
	assert.NoError(err)
}
