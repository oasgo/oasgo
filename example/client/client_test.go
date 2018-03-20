package client

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPets(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/pets", r.URL.EscapedPath())

		resp, _ := ioutil.ReadFile("./testdata/pets.json")
		w.Write(resp)
	}))
	defer s.Close()

	tag := "Good boy"

	c, _ := NewHTTPSwaggerPetstoreClient(s.URL)
	p := Pets{
		Pet{
			ID:   1,
			Name: "Doge",
			Tag:  tag,
		},
	}

	var res Pets
	c.ListPets(&res, nil, 1)
	assert.Equal(t, p, res)
}

func TestShowPetById(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/pets/1", r.URL.EscapedPath())

		resp, _ := ioutil.ReadFile("./testdata/pet.json")
		w.Write(resp)
	}))
	defer s.Close()

	c, _ := NewHTTPSwaggerPetstoreClient(s.URL)
	tag := "Good boy"
	p := Pet{
		ID:   1,
		Name: "Doge",
		Tag:  tag,
	}

	var res Pet
	c.ShowPetById(&res, "1")

	assert.Equal(t, p, res)
}
