package az

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type result struct {
	notDone bool
	nextErr error
}

func (r result) NotDone() bool { return r.notDone }
func (r result) Next() error   { return r.nextErr }

func TestPaginate(t *testing.T) {
	e := errors.New("err")

	pg := Paginate(result{}, e)
	assert.False(t, pg.Next())
	assert.Equal(t, e, pg.Err)

	pg = Paginate(result{}, nil)
	assert.False(t, pg.Next())
	assert.Nil(t, pg.Err)

	pg = Paginate(result{notDone: true, nextErr: e}, nil)
	assert.True(t, pg.Next())
	assert.Nil(t, pg.Err)
	assert.False(t, pg.Next())
	assert.Equal(t, e, pg.Err)

	r := result{notDone: true}
	pg = Paginate(&r, nil)
	assert.True(t, pg.Next())
	assert.Nil(t, pg.Err)
	assert.True(t, pg.Next())
	assert.Nil(t, pg.Err)
	r.notDone = false
	assert.False(t, pg.Next())
	assert.Nil(t, pg.Err)
}
