/*
* This file autogenerated;
*
* DO NOT EDIT
*
* Swagger Petstore
* Version: 1.0.0
*
 */

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	parameterIsMissingErr   = "parameter %s is missing"
	invalidBodyErr          = "invalid body"
	invalidParameterTypeErr = "invalid type of %s: %s"
)

type (
	MissingParameterError struct {
		field string
	}

	InvalidParameterTypeError struct {
		field    string
		original error
	}

	InvalidBodyError struct {
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

func (e *MissingParameterError) Error() string {
	return fmt.Sprintf(parameterIsMissingErr, e.field)
}

func (e *InvalidParameterTypeError) Error() string {
	return fmt.Sprintf(invalidParameterTypeErr, e.field, e.original.Error())
}

func (e *InvalidBodyError) Error() string {
	return invalidBodyErr
}

func ListPets(r *http.Request) (limit *int64, fancyQueryArg int64, err error) {
	value := r.URL.Query().Get("limit")
	*limit, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		err = &InvalidParameterTypeError{
			field:    "limit",
			original: err,
		}
		return
	}
	value = r.URL.Query().Get("fancy_query_arg")
	if value == "" {
		err = &MissingParameterError{field: "fancy_query_arg"}
		return
	}
	fancyQueryArg, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		err = &InvalidParameterTypeError{
			field:    "fancy_query_arg",
			original: err,
		}
		return
	}
	return
}
func CreatePet(r *http.Request) (body Pet, err error) {
	var bs []byte
	if bs, err = ioutil.ReadAll(r.Body); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &body); err != nil {
		err = &InvalidBodyError{}
		return
	}
	return
}
func ShowPetById(r *http.Request) (petId string, err error) {
	value := r.URL.Query().Get("petId")
	if value == "" {
		err = &MissingParameterError{field: "petId"}
		return
	}
	petId = value
	return
}
