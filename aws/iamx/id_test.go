package iamx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntity(t *testing.T) {
	tests := []Entity{
		Group,
		InstanceProfile,
		ManagedPolicy,
		PolicyVersion,
		Role,
		Root,
		ServerCert,
		TempKey,
		User,
		UserKey,
	}
	for _, e := range tests {
		for _, id := range []string{string(e), string(e + "X")} {
			assert.Equal(t, string(e), string(Type(id)), "id=%s", id)
			assert.True(t, Is(id, e), "id=%s", id)
		}
	}
	assert.Equal(t, "", string(Type("")))
	assert.Equal(t, "", string(Type("AIDB")))
}
