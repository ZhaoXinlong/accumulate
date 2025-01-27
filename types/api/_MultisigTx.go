package api

import (
	"bytes"
	"fmt"

	"github.com/AccumulateNetwork/accumulate/smt/common"
	"github.com/AccumulateNetwork/accumulate/types"
)

type MultiSigTx struct {
	TxHash types.Bytes32 `json:"hash" form:"url" query:"url" validate:"required"`
}

func (m *MultiSigTx) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(common.Uint64Bytes(types.TxMultisigTx.AsUint64()))
	buf.Write(m.TxHash[:])

	return buf.Bytes(), nil
}

func (m *MultiSigTx) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if rErr := recover(); rErr != nil {
			err = fmt.Errorf("insufficent data to unmarshal MultiSigTx %v", rErr)
		}
	}()

	txType, data := common.BytesUint64(data)
	if txType != uint64(types.TxMultisigTx) {
		return fmt.Errorf("attempting to unmarshal incompatible type")
	}
	copy(m.TxHash[:], data)
	return nil
}
