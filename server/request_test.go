package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/go-chi/chi"
	"github.com/kochman/hotshots/config"
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
func (db *MockDB) Count(data interface{}) (int, error) {
	args := db.Called(data)
	return args.Int(0), args.Error(1)
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

func (db *MockDB) Select(matchers ...q.Matcher) storm.Query {
	args := db.Called(matchers)
	return args.Get(0).(storm.Query)
}

func (db *MockDB) Update(data interface{}) error {
	args := db.Called(data)
	return args.Error(0)
}

func (db *MockDB) UpdateField(data interface{}, fieldName string, value interface{}) error {
	args := db.Called(data, fieldName, value)
	return args.Error(0)
}

func ImageAsBytes(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func MockPhotoCtx(photo Photo) *http.Request {
	r := httptest.NewRequest("", "/", new(bytes.Reader))
	ctx := context.WithValue(r.Context(), "photo", photo)
	return r.WithContext(ctx)
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

	WriteJsonResponse(TestStruct{"fuck"}, 200, w)

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

func prepareMockServer(t *testing.T) (Server, *MockDB, *MockQuery) {
	dir, err := ioutil.TempDir("", "hotshot")
	require.Nil(t, err)

	err = os.Mkdir(path.Join(dir, "img"), 755)
	require.Nil(t, err)

	err = os.Mkdir(path.Join(dir, "conf.d"), 755)
	require.Nil(t, err)

	cfg := config.Config{
		ListenURL:       "localhost:8000",
		PhotosDirectory: dir,
	}
	db := MockDB{}
	qu := MockQuery{}

	qu.On("Skip", mock.Anything).Return(&qu)
	qu.On("Limit", mock.Anything).Return(&qu)
	qu.On("OrderBy", mock.Anything).Return(&qu)

	return Server{
		cfg:     cfg,
		db:      &db,
		handler: nil,
		timeout: 2 * time.Minute,
	}, &db, &qu
}

func TestGetPhotoFromDatabase(t *testing.T) {
	s, db, _ := prepareMockServer(t)

	db.On("One", "ID", "valid", mock.Anything).Run(func(args mock.Arguments) {
		p := args.Get(2).(*Photo)
		p.ID = "1234"
	}).Return(nil)
	db.On("One", "ID", "invalid", mock.Anything).Return(errors.New(""))

	/* If the one query succeeds */
	p, err := s.GetPhotoFromDatabase("valid")
	require.Nil(t, err)
	assert.EqualValues(t, p.ID, "1234")

	/* If the one query fails */
	p, err = s.GetPhotoFromDatabase("invalid")
	require.NotNil(t, err)
	require.EqualValues(t, p, Photo{})

	db.AssertNumberOfCalls(t, "One", 2)
}

func TestPhotoCtx(t *testing.T) {
	s, db, _ := prepareMockServer(t)

	db.On("One", "ID", "valid", mock.Anything).Run(func(args mock.Arguments) {
		p := args.Get(2).(*Photo)
		p.ID = "1234"
	}).Return(nil)
	db.On("One", "ID", "invalid", mock.Anything).Return(errors.New(""))

	r := chi.NewRouter()

	r.Route("/{pid}", func(r chi.Router) {
		r.Use(s.PhotoCtx)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			photo := r.Context().Value("photo").(Photo)
			require.EqualValues(t, photo.ID, "1234")
			w.Write(nil)
		})
	})

	server := httptest.NewServer(r)
	defer server.Close()

	res, _ := http.Get(server.URL + "/valid")
	require.EqualValues(t, res.StatusCode, 200)

	res, _ = http.Get(server.URL + "/invalid")
	require.EqualValues(t, res.StatusCode, 404)

	db.AssertNumberOfCalls(t, "One", 2)

}

func TestGetPhotos(t *testing.T) {
	s, db, qu := prepareMockServer(t)

	db.On("Select", mock.Anything, mock.Anything).Return(qu)
	qu.On("Find", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		p := args.Get(0).(*[]Photo)
		*p = append(*p, Photo{ID: "1234"})
		*p = append(*p, Photo{ID: "5678"})
	}).Return(nil)

	r := httptest.NewRequest("", "/", new(bytes.Reader))
	w := httptest.NewRecorder()

	s.GetPhotos(w, r)
	qu.AssertNumberOfCalls(t, "Find", 1)
	require.EqualValues(t, w.Code, 200)

	var v GetPhotosResponse
	b, _ := ioutil.ReadAll(w.Body)
	err := json.Unmarshal(b, &v)
	require.Nil(t, err)
	assert.True(t, v.Success)
	assert.Len(t, v.Photos, 2)
	assert.Contains(t, v.Photos, Photo{ID: "1234"})
	assert.Contains(t, v.Photos, Photo{ID: "5678"})
}

func TestGetPhotoIDs(t *testing.T) {
	s, db, qu := prepareMockServer(t)

	db.On("Select", mock.Anything, mock.Anything).Return(qu)
	qu.On("Find", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		p := args.Get(0).(*[]Photo)
		*p = append(*p, Photo{ID: "1234"})
		*p = append(*p, Photo{ID: "5678"})
	}).Return(nil)

	r := httptest.NewRequest("", "/", new(bytes.Reader))
	w := httptest.NewRecorder()

	s.GetPhotoIDs(w, r)
	qu.AssertNumberOfCalls(t, "Find", 1)
	require.EqualValues(t, w.Code, 200)

	var v GetPhotoIDsResponse
	b, _ := ioutil.ReadAll(w.Body)
	err := json.Unmarshal(b, &v)
	require.Nil(t, err)
	assert.True(t, v.Success)
	assert.Len(t, v.IDs, 2)
	assert.Contains(t, v.IDs, "1234")
	assert.Contains(t, v.IDs, "5678")
}

func TestPostPhoto(t *testing.T) {
	s, _, _ := prepareMockServer(t)

	// Photo value not filled
	r := httptest.NewRequest("", "/", new(bytes.Reader))
	w := httptest.NewRecorder()

	s.PostPhoto(w, r)

	var v ErrorResponse
	b, _ := ioutil.ReadAll(w.Body)
	err := json.Unmarshal(b, &v)
	require.Nil(t, err)
	assert.EqualValues(t, w.Code, 400)
	assert.EqualValues(t, v, ErrorResponse{Success: false, Error: "unable to parse form value 'photo'"})

	// TODO: Photo already in DB + No overwrite Flag

	// TODO: Photo already in DB + Overwrite Flag

	// TODO: Photo Processing Failed
	// TODO: Photo Processing Success
}

func TestGetPhotoMetadata(t *testing.T) {
	s, _, _ := prepareMockServer(t)

	r := MockPhotoCtx(Photo{ID: "1234"})
	w := httptest.NewRecorder()

	s.GetPhotoMetadata(w, r)

	require.EqualValues(t, w.Code, 200)

	var v GetPhotoMetadataResponse
	b, _ := ioutil.ReadAll(w.Body)
	err := json.Unmarshal(b, &v)
	require.Nil(t, err)
	assert.EqualValues(t, v.Photo, Photo{ID: "1234"})
}

func TestGetImage(t *testing.T) {
	s, _, _ := prepareMockServer(t)

	func() {
		validImage, err := os.Open("_testdata/validFile.jpg")
		require.Nil(t, err)
		defer validImage.Close()

		f, err := os.Create(path.Join(s.cfg.ImgFolder(), "1234.jpg"))
		require.Nil(t, err)
		defer f.Close()

		io.Copy(f, validImage)
	}()

	// Processing not done/failed
	r := MockPhotoCtx(Photo{Status: ProcessingFailed})
	w := httptest.NewRecorder()

	s.GetImage("%s.jpg", w, r)
	require.EqualValues(t, w.Code, 400)

	var v ErrorResponse
	b, _ := ioutil.ReadAll(w.Body)
	err := json.Unmarshal(b, &v)
	require.Nil(t, err)
	assert.False(t, v.Success)
	assert.EqualValues(t, v.Error, "photo not processed")

	// Image not accessible
	r = MockPhotoCtx(Photo{Status: ProcessingSucceeded})
	w = httptest.NewRecorder()

	s.GetImage("%s.jpg", w, r)
	require.EqualValues(t, w.Code, 500)

	b, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(b, &v)
	require.Nil(t, err)
	assert.False(t, v.Success)
	assert.EqualValues(t, v.Error, "unable to access internal storage of image")

	// Processing not done/failed
	r = MockPhotoCtx(Photo{ID: "1234", Status: ProcessingSucceeded})
	w = httptest.NewRecorder()

	s.GetImage("%s.jpg", w, r)
	require.EqualValues(t, w.Code, 200)
	assert.EqualValues(t, w.Header().Get("Content-Type"), "image/jpeg")
}
