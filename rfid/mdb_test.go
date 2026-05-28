package rfid

import (
	"fmt"
	"github.com/itchyny/base58-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBase58(t *testing.T) {
	t.Run("basic 0 is 1", func(t *testing.T) {
		exId := MainCollectionId([8]byte{0, 0, 0, 0, 0, 0, 0, 0})
		assert.Equal(t, Base58Str("0"), exId.AsBase58())
	})
	t.Run("basic 0 is 0", func(t *testing.T) {
		decoded, err := base58.BitcoinEncoding.Decode([]byte("1"))
		assert.NoError(t, err)
		assert.Equal(t, "0", string(decoded))
	})
	for _, size := range []int{8, 12} {
		sizStr := fmt.Sprintf("%d", size)
		effBytes := randomRFID(size)
		b58, err := Base2BytesToBase58(effBytes)
		assert.NoError(t, err, "Error converting "+sizStr+" to base58")
		effBytes2, err := b58.Base2Bytes()
		assert.NoError(t, err, "Error converting "+sizStr+" back to littleEndian")
		b58b, err := Base2BytesToBase58(effBytes2)
		assert.NoError(t, err, "Error converting "+sizStr+" to base58 a second time")
		assert.Equal(t, string(effBytes), string(effBytes2), "littleEndian strings should match for "+sizStr)
		assert.Equal(t, string(b58), string(b58b), "base58 strings should match for "+sizStr)
		assert.NotEqual(t, string(effBytes), string(b58b), "base58 and efficient strings should not match for "+sizStr)
	}
}
