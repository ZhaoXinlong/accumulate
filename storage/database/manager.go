package database

import (
	"crypto/sha256"
	"fmt"

	"github.com/AccumulateNetwork/SMT/storage"
	"github.com/AccumulateNetwork/SMT/storage/badger"
	"github.com/AccumulateNetwork/SMT/storage/memory"
)

type Manager struct {
	DB      storage.KeyValueDB // Underlying database implementation
	Buckets map[string]byte    // one byte to indicate a bucket
	Labels  map[string]byte    // one byte to indicate a label
	TXList  TXList             // Transaction List
	Count   int64              // The number of elements in this database
}

// Init
// Initialize the Manager with a specified underlying database
func (m *Manager) Init(database, filename string) error {
	// Set up Buckets for use by the Stateful Merkle Trees
	m.Buckets = make(map[string]byte)
	//  Bucket and Label use by Stateful Merkle Trees
	//  Bucket             --          key/value
	//  ElementIndex       -- element hash / element index
	//  States             -- element index / merkle state
	//  NextElement        -- element index / next element
	//  Element            -- "Count" / highest element

	// element hash / element index
	m.AddBucket("ElementIndex")
	// element index / element state
	m.AddBucket("States")
	// element index / next element
	m.AddBucket("NextElement")
	// count of element in the database
	m.AddBucket("Element")

	// match with a supported database
	switch database {
	case "badger": // Badger
		m.DB = new(badger.DB)
		if err := m.DB.InitDB(filename); err != nil {
			return err
		}
	case "memory": // DB databases can't fail
		m.DB = new(memory.DB)
		_ = m.DB.InitDB(filename)
	}
	m.Count = m.GetCount()
	return nil
}

// GetCount
// The number of elements as recorded in the Database.  Note that this may differ from
// the count in the the actual Merkle Tree in memory due to batching transactions
func (m *Manager) GetCount() int64 {
	// Look and see if there is any element count recorded
	data := m.DB.Get(m.GetKey("Element", "", []byte("Count")))
	if data == nil { // If not, nothing is there
		return 0
	}
	v, _ := storage.BytesInt64(data) // Return the recorded count
	return v
}

// Close
// Do any cleanup required to close the manager
func (m *Manager) Close() {
	m.DB.Close()
}

// AddBucket
// Add a bucket to be used in the database.  Initializing a database requires
// that buckets and labels be added in the correct order.
func (m *Manager) AddBucket(bucket string) {
	if _, ok := m.Buckets[bucket]; ok {
		panic(fmt.Sprintf("the bucket '%s' is already defined", bucket))
	}
	idx := len(m.Buckets) + 1
	if idx > 255 {
		panic("too many buckets")
	}
	m.Buckets[bucket] = byte(idx)
}

// AddLabel
// Add a Label to be used in the database.  Initializing a database requires
// that buckets and labels be added in the correct order.
func (m *Manager) AddLabel(label string) {
	if _, ok := m.Labels[label]; ok {
		panic(fmt.Sprintf("the label '%s' is already defined", label))
	}
	idx := len(m.Labels) + 1
	if idx > 255 {
		panic("too many labels")
	}
	m.Buckets[label] = byte(idx)
}

// GetKey
// Given a Bucket Name, a Label name, and a key, GetKey returns a single
// key to be used in a key/value database
func (m *Manager) GetKey(Bucket, Label string, key []byte) (DBKey [storage.KeyLength]byte) {
	var ok bool
	if _, ok = m.Buckets[Bucket]; !ok { //                                 Is the bucket defined?
		panic(fmt.Sprintf("bucket %s undefined or invalid", Bucket)) //      Panic if not
	}
	if _, ok = m.Labels[Label]; len(Label) > 0 && !ok { //                 If a label is specified, is it defined?
		panic(fmt.Sprintf("label %s undefined or invalid", Label)) //        Panic if not.
	}
	DBKey = sha256.Sum256(key) //          To get a fixed length key, we hash the given key.
	//                                     Because a hash is very secure, losing two bytes won't hurt anything
	DBKey[0] = m.Buckets[Bucket] //        Replace the first byte with the bucket index
	DBKey[1] = 0                 //        Assume no label (0 -- means no label
	if len(Label) > 0 {          //        But if a label is specified, then
		DBKey[1] = m.Labels[Label] //        set its byte value.
	}

	return DBKey
}

// Put
// Put a []byte value into the underlying database
func (m *Manager) Put(Bucket, Label string, key []byte, value []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	k := m.GetKey(Bucket, Label, key)
	err = m.DB.Put(k, value)
	return err
}

// PutString
// Put a String value into the underlying database
func (m *Manager) PutString(Bucket, Label string, key []byte, value string) error {
	return m.Put(Bucket, Label, key, []byte(value))
}

// PutInt64
// Put a int64 value into the underlying database
func (m *Manager) PutInt64(Bucket, Label string, key []byte, value int64) error {
	return m.Put(Bucket, Label, key, storage.Int64Bytes(value))
}

// Get
// Get a []byte value from the underlying database.  Returns a nil if not found, or on an error
func (m *Manager) Get(Bucket, Label string, key []byte) (value []byte) {
	m.EndBatch()
	return m.DB.Get(m.GetKey(Bucket, Label, key))
}

// GetString
// Get a string value from the underlying database.  Returns a nil if not found, or on an error
func (m *Manager) GetString(Bucket, Label string, key []byte) (value string) {
	return string(m.DB.Get(m.GetKey(Bucket, Label, key)))
}

// GetInt64
// Get a string value from the underlying database.  Returns a nil if not found, or on an error
func (m *Manager) GetInt64(Bucket, Label string, key []byte) (value int64) {
	v, _ := storage.BytesInt64(m.DB.Get(m.GetKey(Bucket, Label, key)))
	return v
}

// GetIndex
// Return the int64 value tied to the element hash in the ElementIndex bucket
func (m *Manager) GetIndex(element []byte) int64 {
	data := m.Get("ElementIndex", "", element)
	if data == nil {
		return -1
	}
	v, _ := storage.BytesInt64(data)
	return v
}

// PutBatch
// put the write of a key value into the pending batch.  These will all be written to the
// database together.
func (m *Manager) PutBatch(Bucket, Label string, key []byte, value []byte) error {
	theKey := m.GetKey(Bucket, Label, key)
	return m.TXList.Put(theKey, value)
}

// EndBatch()
// Flush anything in the batch list to the database.
func (m *Manager) EndBatch() {
	if len(m.TXList.List) == 0 { // If there is nothing to do, do nothing
		return
	}
	if err := m.DB.PutBatch(m.TXList.List); err != nil {
		panic("batch failed to persist to the database")
	}
	m.TXList.List = m.TXList.List[:0] // Reset the List to allow it to be reused
}

func (m *Manager) BeginBatch() {
	m.TXList.List = m.TXList.List[:0]
}
