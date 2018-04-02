/*
* This file autogenerated;
*
* DO NOT EDIT
*
* Swagger Petstore
* Version: 1.0.0
*
 */

package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var _ SwaggerPetstore = new(HTTPSwaggerPetstoreClient)

type (
	SwaggerPetstore interface {
		ListPets(res interface{}, limit *int64, fancyQueryArg int64) (*http.Response, error)
		CreatePet(res interface{}, body Pet) (*http.Response, error)
		ShowPetById(res interface{}, petId string) (*http.Response, error)
	}

	HTTPSwaggerPetstoreClient struct {
		URL  *url.URL
		HTTP *http.Client
	}

	Error struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	}

	Pet struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}

	Pets []Pet

	PetsLala []int64
)

func NewHTTPSwaggerPetstoreClient(host string) (*HTTPSwaggerPetstoreClient, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	return &HTTPSwaggerPetstoreClient{
		URL:  u,
		HTTP: &http.Client{},
	}, nil
}

// ListPets
func (c HTTPSwaggerPetstoreClient) ListPets(res interface{}, limit *int64, fancyQueryArg int64) (*http.Response, error) {
	u := *c.URL
	u.Path = "/pets"

	u.Path = strings.NewReplacer().Replace(u.Path)

	q := u.Query()
	if limit != nil {
		q.Set("limit", strconv.FormatInt(*limit, 10))
	}
	q.Set("fancy_query_arg", strconv.FormatInt(fancyQueryArg, 10))

	u.RawQuery = q.Encode()

	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r, ok := res.(*string); ok {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return resp, err
		}

		*r = string(body)
		return resp, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(res); err != nil {
		return resp, err
	}

	return resp, nil
}

// CreatePet
func (c HTTPSwaggerPetstoreClient) CreatePet(res interface{}, body Pet) (*http.Response, error) {
	u := *c.URL
	u.Path = "/pets"

	u.Path = strings.NewReplacer().Replace(u.Path)

	q := u.Query()
	u.RawQuery = q.Encode()

	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r, ok := res.(*string); ok {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return resp, err
		}

		*r = string(body)
		return resp, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(res); err != nil {
		return resp, err
	}

	return resp, nil
}

// ShowPetById
func (c HTTPSwaggerPetstoreClient) ShowPetById(res interface{}, petId string) (*http.Response, error) {
	u := *c.URL
	u.Path = "/pets/{petId}"

	u.Path = strings.NewReplacer(
		"{petId}", petId,
	).Replace(u.Path)

	q := u.Query()
	u.RawQuery = q.Encode()

	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r, ok := res.(*string); ok {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return resp, err
		}

		*r = string(body)
		return resp, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(res); err != nil {
		return resp, err
	}

	return resp, nil
}
