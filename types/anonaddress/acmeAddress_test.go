package types

import (
	"crypto/sha256"
	"testing"
)

// GenerateAcmeAddress is still referenced, so testing it seems reasonable
func TestGenerateAcmeAddress(t *testing.T) {

	pub := sha256.Sum256([]byte{1, 2, 3, 3, 2, 1}) // can't tell a public key from a hash.
	var list []string
	for i := 0; i < 100; i++ {
		adr := GenerateAcmeAddress(pub[:])
		// fmt.Printf("%X ", pub)
		list = append(list, adr)
		pub = sha256.Sum256(pub[:])
		//println(adr, " ", len(adr))
	}
	pub = sha256.Sum256([]byte{1, 2, 3, 3, 2, 1}) // can't tell a public key from a hash.
	for _, adr := range list {
		adr2 := GenerateAcmeAddress(pub[:])
		if adr != adr2 {
			t.Error("adr != adr2 should never happen")
		}
		pub = sha256.Sum256(pub[:])
	}

}
