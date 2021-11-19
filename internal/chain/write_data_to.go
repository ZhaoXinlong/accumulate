package chain

import (
	"errors"
	"fmt"

	"github.com/AccumulateNetwork/accumulate/protocol"
	"github.com/AccumulateNetwork/accumulate/types"
	"github.com/AccumulateNetwork/accumulate/types/api/transactions"
)

type WriteDataTo struct{}

func (WriteDataTo) Type() types.TxType { return types.TxTypeWriteDataTo }

func (WriteDataTo) Validate(st *StateManager, tx *transactions.GenTransaction) error {
	body := new(protocol.WriteDataTo)
	err := tx.As(body)
	if err != nil {
		return fmt.Errorf("invalid payload: %v", err)
	}

	return errors.New("not implemented") // TODO
}