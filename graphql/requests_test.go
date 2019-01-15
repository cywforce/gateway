package graphql

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestNetworkQueryer_sendsQueries(t *testing.T) {
	// build a query to test should be equivalent to
	// targetQueryBody := `
	// 	{
	// 		hello(world: "hello") {
	// 			world
	// 		}
	// 	}
	// `

	// the result we expect back
	expected := map[string]interface{}{
		"data": map[string]interface{}{
			"foo": "bar",
		},
	}

	// create a http client that responds with a known body and verifies the incoming query
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			// serialize the json we want to send back
			result, err := json.Marshal(expected)
			// if something went wrong
			if err != nil {
				return &http.Response{
					StatusCode: 500,
					Body:       ioutil.NopCloser(bytes.NewBufferString("Something went wrong")),
					Header:     make(http.Header),
				}
			}

			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBuffer(result)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		}),
	}

	// the corresponding query document
	query := `
		{
			hello(world: "hello") {
				world
			}
		}
	`

	queryer := &NetworkQueryer{
		URL:    "hello",
		Client: client,
	}

	// get the response of the query
	result := map[string]interface{}{}
	err := queryer.Query(&QueryInput{Query: query}, &result)
	if err != nil {
		t.Error(err)
		return
	}
	if result == nil {
		t.Error("Did not get a result back")
		return
	}

	// make sure we got what we expected
	assert.Equal(t, expected["data"], result)
}

func TestNetworkQueryer_handlesErrorResponse(t *testing.T) {
	// build a query to test should be equivalent to
	// targetQueryBody := `
	// 	{
	// 		hello(world: "hello") {
	// 			world
	// 		}
	// 	}
	//

	// create a http client that responds with a known body and verifies the incoming query
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			// serialize the json we want to send back
			result, err := json.Marshal(map[string]interface{}{
				"data": nil,
				"errors": []map[string]interface{}{
					map[string]interface{}{
						"message": "message",
					},
				},
			})
			// if something went wrong
			if err != nil {
				return &http.Response{
					StatusCode: 500,
					Body:       ioutil.NopCloser(bytes.NewBufferString("Something went wrong")),
					Header:     make(http.Header),
				}
			}

			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBuffer(result)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		}),
	}

	// the corresponding query document
	query := `
		{
			hello(world: "hello") {
				world
			}
		}
	`

	queryer := &NetworkQueryer{
		URL:    "hello",
		Client: client,
	}

	// get the response of the query
	result := map[string]interface{}{}
	err := queryer.Query(&QueryInput{Query: query}, &result)

	//
	assert.NotNil(t, err)
	assert.Equal(t, "message", err.Error())
}

func TestNetworkQueryer_respondsWithErr(t *testing.T) {
	// create a http client that responds with a known body and verifies the incoming query
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			// send an error back
			return &http.Response{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewBufferString("Something went wrong")),
				Header:     make(http.Header),
			}
		}),
	}

	// the corresponding query document
	query := `
		{
			hello
		}
	`

	queryer := &NetworkQueryer{
		URL:    "hello",
		Client: client,
	}

	// get the response of the query
	var result interface{}
	err := queryer.Query(&QueryInput{Query: query}, result)
	if err == nil {
		t.Error("Did not receive an error")
		return
	}
}

func TestNewNetworkQueryer(t *testing.T) {
	// make sure that create a new query renderer saves the right URL
	assert.Equal(t, "foo", NewNetworkQueryer("foo").URL)
}

func TestQueryerFunc(t *testing.T) {
	expected := map[string]interface{}{"hello": "world"}

	queryer := QueryerFunc{
		func(*QueryInput) (interface{}, error) {
			return expected, nil
		},
	}

	// a place to write the result
	result := map[string]interface{}{}

	queryer.Query(&QueryInput{}, &result)

	// make sure we copied the right result
	assert.Equal(t, expected, result)
}
