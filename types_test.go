package go_parquet

import (
	"bytes"
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildRandArray(count int, fn func() interface{}) []interface{} {
	ret := make([]interface{}, count)
	for i := range ret {
		ret[i] = fn()
	}

	return ret
}

type testFixtures struct {
	name string
	enc  valuesEncoder
	dec  valuesDecoder
	rand func() interface{}
}

var (
	tests = []testFixtures{
		{
			name: "Int32Plain",
			enc:  &int32PlainEncoder{},
			dec:  &int32PlainDecoder{},
			rand: func() interface{} {
				return int32(rand.Int())
			},
		},
		{
			name: "Int64Plain",
			enc:  &int64PlainEncoder{},
			dec:  &int64PlainDecoder{},
			rand: func() interface{} {
				return rand.Int63()
			},
		},
		{
			name: "Int96Plain",
			enc:  &int96PlainEncoder{},
			dec:  &int96PlainDecoder{},
			rand: func() interface{} {
				var data Int96
				for i := 0; i < 12; i++ {
					data[i] = byte(rand.Intn(256))
				}

				return data
			},
		},
		{
			name: "DoublePlain",
			enc:  &doublePlainEncoder{},
			dec:  &doublePlainDecoder{},
			rand: func() interface{} {
				return rand.Float64()
			},
		},
		{
			name: "FloatPlain",
			enc:  &floatPlainEncoder{},
			dec:  &floatPlainDecoder{},
			rand: func() interface{} {
				return rand.Float32()
			},
		},
	}
)

func TestTypes(t *testing.T) {
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			arr1 := buildRandArray(1000, data.rand)
			arr2 := buildRandArray(1000, data.rand)
			w := &bytes.Buffer{}
			assert.NoError(t, data.enc.init(w))
			assert.NoError(t, data.enc.encodeValues(arr1))
			assert.NoError(t, data.enc.encodeValues(arr2))
			if c, ok := data.enc.(io.Closer); ok {
				assert.NoError(t, c.Close())
			}

			ret := make([]interface{}, 1000)
			r := bytes.NewReader(w.Bytes())
			assert.NoError(t, data.dec.init(r))
			assert.NoError(t, data.dec.decodeValues(ret))
			assert.Equal(t, ret, arr1)
			assert.NoError(t, data.dec.decodeValues(ret))
			assert.Equal(t, ret, arr2)
			// No more data
			assert.Error(t, data.dec.decodeValues(ret))
		})
	}
}
