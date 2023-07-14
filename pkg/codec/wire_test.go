package codec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	var v Version

	t.Run("encoding", func(t *testing.T) {
		ans, err := json.Marshal(v)
		assert.NoError(t, err)
		assert.Equal(t, []byte(`"2.0"`), ans)
	})

	t.Run("decoding", func(t *testing.T) {
		err := json.Unmarshal([]byte(`"2.0"`), &v)
		assert.NoError(t, err)
		err = json.Unmarshal([]byte("not"), &v)
		assert.Error(t, err)
	})
}

func TestID(t *testing.T) {

	var v ID

	t.Run("number", func(t *testing.T) {
		v = NewNumberID(2)
		ans, err := json.Marshal(v)
		assert.NoError(t, err)
		assert.Equal(t, `2`, string(ans))
	})

	t.Run("numberstring", func(t *testing.T) {
		v = NewStringID("2")
		ans, err := json.Marshal(v)
		assert.NoError(t, err)
		assert.Equal(t, `"2"`, string(ans))
	})
	t.Run("string", func(t *testing.T) {
		v = NewStringID("doggo")
		ans, err := json.Marshal(v)
		assert.NoError(t, err)
		assert.Equal(t, `"doggo"`, string(ans))
	})
	t.Run("null", func(t *testing.T) {
		v = NewNullID()
		ans, err := json.Marshal(v)
		assert.NoError(t, err)
		assert.Equal(t, `null`, string(ans))
	})
}
