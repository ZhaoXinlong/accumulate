package managed

import (
	"bytes"
	"crypto/sha256"
	"math"
	"testing"

	"github.com/AccumulateNetwork/SMT/storage/database"
)

func TestAddSalt(t *testing.T) {
	Salt := sha256.Sum256([]byte{1})
	Salt2 := add2Salt(Salt[:], 1)
	if bytes.Equal(Salt[:], Salt2) {
		t.Errorf("These should not be equal \n%x \n%x", Salt, Salt2)
	}
}

func TestIndexing(t *testing.T) {

	const testlen = 1024
	const blocklen = 10

	dbManager := new(database.Manager)
	if err := dbManager.Init("memory", ""); err != nil {
		t.Fatal(err)
	}

	salt := sha256.Sum256([]byte("root"))
	MM1 := NewMerkleManager(dbManager, salt[:], 2)

	// Fill the Merkle Tree with a few hashes
	hash := sha256.Sum256([]byte("start"))
	for i := 0; i < testlen; i++ {
		MM1.AddHash(hash)
		hash = sha256.Sum256(hash[:])
		if (i+1)%blocklen == 0 {
			MM1.SetBlockIndex(int64(i) / blocklen)
		}
	}

	hash = sha256.Sum256([]byte("start"))
	for i := int64(0); i < testlen; i++ {
		if (i+1)%blocklen == 0 {
			bi := new(BlockIndex)
			data := MM1.MainChain.Manager.Get("BlockIndex", "", Int64Bytes(i/blocklen))
			bi.UnMarshal(data)
			if bi.MainIndex != i {
				t.Fatalf("the MainIndex doesn't match v %d i %d",
					bi.MainIndex, i)
			}
			if bi.BlockIndex != i/blocklen {
				t.Fatalf("the BlockIndex doesn't match v %d i/blocklen %d",
					bi.BlockIndex, i/blocklen)
			}
		}
		if v := MM1.GetIndex(hash[:]); v < 0 {
			t.Fatalf("failed to index hash %d", i)
		} else {
			if v != i {
				t.Fatalf("failed to get the right index.  i %d v %d", i, v)
			}
		}
		hash = sha256.Sum256(hash[:])
	}

	MM2 := NewMerkleManager(dbManager, []byte("root"), 2)

	if MM1.MainChain.MS.Count != MM2.MainChain.MS.Count {
		t.Fatal("failed to properly load from a database")
	}

	hash = sha256.Sum256([]byte("start"))
	for i := 0; i < testlen; i++ {
		if v := MM2.GetIndex(hash[:]); v < 0 {
			t.Fatalf("failed to index hash %d", i)
		} else {
			if int(v) != i {
				t.Fatalf("failed to get the right index.  i %d v %d", i, v)
			}
		}
		hash = sha256.Sum256(hash[:])
	}
}

func TestMerkleManager(t *testing.T) {

	const testlen = 1024

	dbManager := new(database.Manager)
	if err := dbManager.Init("memory", ""); err != nil {
		t.Fatal(err)
	}

	MarkPower := int64(2)
	MarkFreq := int64(math.Pow(2, float64(MarkPower)))
	MarkMask := MarkFreq - 1

	// Set up a MM1 that uses a MarkPower of 2
	MM1 := NewMerkleManager(dbManager, []byte("root"), MarkPower)

	if MarkPower != MM1.MarkPower ||
		MarkFreq != MM1.MarkFreq ||
		MarkMask != MM1.MarkMask {
		t.Fatal("Marks were not correctly computed")
	}

	// Fill the Merkle Tree with a few hashes
	hash := sha256.Sum256([]byte("start"))
	for i := 0; i < testlen; i++ {
		MM1.AddHash(hash)
		hash = sha256.Sum256(hash[:])
	}

	if MM1.GetElementCount() != testlen {
		t.Fatal("added elements in merkle tree don't match the number we added")
	}

	dbManager.EndBatch()

	// Check the Indexing
	for i := int64(0); i < testlen; i++ {
		ms := MM1.GetState(i)
		m := MM1.GetNext(i)
		if (i+1)&MarkMask == 0 {
			if ms == nil {
				t.Fatal("should have a state at Mark point - 1 at ", i)
			}
			if m == nil {
				t.Fatal("should have a next element at Mark point - 1 at ", i)
			}
		} else if i&MarkMask == 0 {
			if ms == nil {
				t.Fatal("should have a state at Mark point at ", i)
			}
			if m != nil {
				t.Fatal("should not have a next element at Mark point at ", i)
			}
		} else {
			if ms != nil {
				t.Fatal("should not have a state outside Mark points at ", i)
			}
			if m != nil {
				t.Fatal("should not have a next element outside Mark points at ", i)
			}
		}

	}
}
