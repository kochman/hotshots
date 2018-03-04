package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/asdine/storm/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDB struct {
	mock.Mock
}

func (db *MockDB) All(to interface{}, options ...func(*index.Options)) error {
	args := db.Called(to, options)
	return args.Error(0)
}

func (db *MockDB) DeleteStruct(data interface{}) error {
	args := db.Called(data)
	return args.Error(0)
}

func (db *MockDB) Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error {
	args := db.Called(fieldName, value, to, options)
	return args.Error(0)
}

func (db *MockDB) Init(data interface{}) error {
	args := db.Called(data)
	return args.Error(0)
}

func (db *MockDB) One(fieldName string, value interface{}, to interface{}) error {
	args := db.Called(fieldName, value, to)
	return args.Error(0)
}

func (db *MockDB) Save(data interface{}) error {
	args := db.Called(data)
	return args.Error(0)
}

func (db *MockDB) Update(data interface{}) error {
	args := db.Called(data)
	return args.Error(0)
}

func TestNotFound(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	NotFound(w, req)

	var v NotFoundResponse
	resp := w.Result()

	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	err = json.Unmarshal(b, &v)
	require.Nil(t, err)

	assert.EqualValues(t, w.Code, 404)
	assert.EqualValues(t, false, v.Success)
	assert.EqualValues(t, "requested URI not found", v.Error)
	assert.EqualValues(t, req.RequestURI, v.URI)
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError("fuck", 400, w)

	var v ErrorResponse
	resp := w.Result()

	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	err = json.Unmarshal(b, &v)
	require.Nil(t, err)

	assert.EqualValues(t, w.Code, 400)
	assert.EqualValues(t, false, v.Success)
	assert.EqualValues(t, "fuck", v.Error)
}

func TestWriteJsonResponse(t *testing.T) {
	type TestStruct struct {
		Test string
	}
	/* Test regular type */

	w := httptest.NewRecorder()

	WriteJsonResponse(&TestStruct{"fuck"}, 200, w)

	var v TestStruct
	resp := w.Result()

	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	err = json.Unmarshal(b, &v)
	require.Nil(t, err)

	assert.EqualValues(t, w.Code, 200)
	assert.EqualValues(t, resp.Header.Get("Content-Type"), "application/json")
	assert.EqualValues(t, "fuck", v.Test)

	/* Test type unsupported for marshaling */

	w = httptest.NewRecorder()
	c := make(chan int)
	defer close(c)
	WriteJsonResponse(c, 200, w) // Can't Unmarshal Channels
	resp = w.Result()
	assert.EqualValues(t, w.Code, 500)
}
