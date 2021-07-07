package managed

// BlockIndex
// Holds a mapping of the BlockIndex to the MainIndex and PendingIndex that mark the end of the block
type BlockIndex struct {
	BlockIndex   int64 // index of the block
	MainIndex    int64 // index of the last element in the main chain
	PendingIndex int64 // index of the last element in the Pending chain
}

// Marshal
// serialize a BlockIndex into a slice of data
func (b *BlockIndex) Marshal() (data []byte) {
	data = append(Int64Bytes(b.BlockIndex), Int64Bytes(b.MainIndex)...)
	data = append(Int64Bytes(b.BlockIndex), Int64Bytes(b.PendingIndex)...)
	return data
}

// UnMarshal
// Extract a BlockIndex from a given slice.  Return the remaining slice
func (b *BlockIndex) UnMarshal(data []byte) (newData []byte) {
	b.BlockIndex, data = BytesInt64(data)
	b.MainIndex, data = BytesInt64(data)
	b.PendingIndex, data = BytesInt64(data)
	return data
}
