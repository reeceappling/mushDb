package rfid

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBase58(t *testing.T) {
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
