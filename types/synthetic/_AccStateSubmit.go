package synthetic

import (
	"fmt"

	"github.com/AccumulateNetwork/accumulate/smt/common"
	"github.com/AccumulateNetwork/accumulate/types"
)

// AccStateSubmit structure sends the information needed for the state of a BVC or DC on accumulate.
// If a BVC, then this is used to send information about the state of the BVC to the DC. If a DC, then
// this sends the information about the state of the DC to the BVC's
type AccStateSubmit struct {
	LeaderUrl string
	NetworkId int64
	Height    int64
	BptHash   types.Bytes32
}

// MarshalBinary serializes the AccStateSubmit struct
func (s *AccStateSubmit) MarshalBinary() (data []byte, err error) {
	data = common.Uint64Bytes(uint64(types.TxBvcSubmission))
	data = append(data, common.SliceBytes([]byte(s.LeaderUrl))...)
	data = append(data, common.Int64Bytes(s.NetworkId)...)
	data = append(data, common.Int64Bytes(s.Height)...)
	data = append(data, common.SliceBytes(s.BptHash[:])...)
	return data, nil
}

// UnmarshalBinary deserializes the AccStateSubmit struct
func (s *AccStateSubmit) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if rErr := recover(); rErr != nil {
			err = fmt.Errorf("error marshaling Bvc State Submit, %v", rErr)
		}
	}()

	var typeId uint64
	typeId, data = common.BytesUint64(data)
	if types.TxType(typeId) != types.TxBvcSubmission {
		return fmt.Errorf("invalid type, received %s(%d) but expected %s", types.TxType(typeId).Name(), typeId,
			types.TxBvcSubmission.Name())
	}

	leaderUrl, data := common.BytesSlice(data)
	s.LeaderUrl = string(leaderUrl)
	s.NetworkId, data = common.BytesInt64(data)
	s.Height, data = common.BytesInt64(data)
	bptHash, data := common.BytesSlice(data)
	copy(s.BptHash[:], bptHash)

	return nil
}
