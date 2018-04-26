package server

import (
	"github.com/asdine/storm"
	"github.com/stretchr/testify/mock"
)

type MockQuery struct {
	mock.Mock
}

func (q *MockQuery) Skip(a int) storm.Query {
	args := q.Called(a)
	return args.Get(0).(storm.Query)
}

func (q *MockQuery) Limit(a int) storm.Query {
	args := q.Called(a)
	return args.Get(0).(storm.Query)
}

func (q *MockQuery) OrderBy(a ...string) storm.Query {
	args := q.Called(a)
	return args.Get(0).(storm.Query)
}

func (q *MockQuery) Reverse() storm.Query {
	args := q.Called()
	return args.Get(0).(storm.Query)

}

func (q *MockQuery) Bucket(a string) storm.Query {
	args := q.Called(a)
	return args.Get(0).(storm.Query)

}

func (q *MockQuery) Find(a interface{}) error {
	args := q.Called(a)
	return args.Error(0)
}

func (q *MockQuery) First(a interface{}) error {
	args := q.Called(a)
	return args.Error(0)
}

func (q *MockQuery) Delete(a interface{}) error {
	args := q.Called(a)
	return args.Error(0)
}

func (q *MockQuery) Count(a interface{}) (int, error) {
	args := q.Called(a)
	return args.Int(0), args.Error(1)
}

func (q *MockQuery) Raw() ([][]byte, error) {
	args := q.Called()
	return args.Get(0).([][]byte), args.Error(1)
}

func (q *MockQuery) RawEach(a func([]byte, []byte) error) error {
	args := q.Called(a)
	return args.Error(0)
}

func (q *MockQuery) Each(a interface{}, b func(interface{}) error) error {
	args := q.Called(a, b)
	return args.Error(0)
}
