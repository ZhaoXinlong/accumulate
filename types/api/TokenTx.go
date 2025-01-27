package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/AccumulateNetwork/accumulate/smt/common"
	"github.com/AccumulateNetwork/accumulate/types"
)

const MaxTokenTxOutputs = 100

type SendTokens struct {
	Hash types.Bytes32     `json:"hash,omitempty" form:"hash" query:"hash" validate:"required"`
	From types.UrlChain    `json:"from" form:"from" query:"from" validate:"required"`
	To   []*TokenRecipient `json:"to" form:"to" query:"to" validate:"required"`
	Meta json.RawMessage   `json:"meta,omitempty" form:"meta" query:"meta" validate:"required"`
}

type TokenTxRequest struct {
	Hash types.Bytes32 `json:"hash" form:"hash" query:"hash" validate:"required"`
}

//TokenRecipient is the structure for the output.  Only handles 64 bit amounts at this time.
type TokenRecipient struct {
	URL    types.UrlChain `json:"url" form:"url" query:"url" validate:"required"`
	Amount uint64         `json:"amount" form:"amount" query:"amount" validate:"gt=0"`
}

func NewTokenTx(from types.String, to ...*TokenRecipient) *SendTokens {
	tx := &SendTokens{}
	tx.From = types.UrlChain{String: from}
	tx.To = to
	return tx
}

func NewTokenTxOutput(url types.String, amount uint64) *TokenRecipient {
	txo := new(TokenRecipient)
	txo.URL = types.UrlChain{String: url}
	txo.Amount = amount
	return txo
}

func (*SendTokens) GetType() types.TxType { return types.TxTypeSendTokens }

func (t *SendTokens) AddToAccount(toUrl types.String, amt uint64) {
	txOut := TokenRecipient{types.UrlChain{String: toUrl}, amt}
	t.To = append(t.To, &txOut)
}

func (t *SendTokens) SetMetadata(md *json.RawMessage) error {
	if md == nil {
		return fmt.Errorf("invalid metadata")
	}
	copy(t.Meta[:], (*md)[:])
	return nil
}

// Equal
// returns true if t == t2, otherwise return false.  Comparing t with t2, if
// any runtime error occurs we return false
func (t *SendTokens) Equal(t2 *SendTokens) (ret bool) {
	defer func() { //                      ret will default to false, so if any error occurs
	}() //                                 we will return false as long as we catch any errors
	if t.From != t2.From { //  Make sure accountURLs are the same
		return false
	}
	tLen := len(t.To)                                               // Get our len
	if tLen != len(t2.To) || tLen < 1 || tLen > MaxTokenTxOutputs { // Make sure len is in range and same as t2
		return false //                                       If anything is different, function is false.
	}
	for i := range t.To {
		if !t.To[i].Equal(t2.To[i]) {
			return false
		}
	}
	return true //                                           Only at the very end after all is done we return true
}

// MarshalBinary serialize the token transaction
func (t *SendTokens) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(common.Uint64Bytes(types.TxTypeSendTokens.ID()))

	data, err := t.From.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("error marshaling From %s,%v", t.From, err)
	}
	buffer.Write(data)

	numOutputs := uint64(len(t.To))
	if numOutputs > MaxTokenTxOutputs {
		return nil, fmt.Errorf("too many outputs for token transaction, please specify between 1 and %d outputs", MaxTokenTxOutputs)
	}
	if numOutputs < 1 {
		return nil, fmt.Errorf("insufficient token transaction outputs, please specify between 1 and %d outputs", MaxTokenTxOutputs)
	}
	buffer.Write(common.Uint64Bytes(numOutputs))
	for i, v := range t.To {
		data, err = v.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("error marshaling To[%d] %s,%v", i, v.URL, err)
		}
		buffer.Write(data)
	}

	a := types.Bytes(t.Meta)
	if a != nil {
		data, err = a.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("error marshalling meta data, %v", err)
		}
		buffer.Write(data)
	}

	return buffer.Bytes(), nil
}

// UnmarshalBinary deserialize the token transaction
func (t *SendTokens) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error marshaling TokenTx State %v", r)
		}
	}()

	txType, data := common.BytesUint64(data) // get the type

	if txType != types.TxTypeSendTokens.ID() {
		return fmt.Errorf("invalid transaction type, expecting TokenTx")
	}

	err = t.From.UnmarshalBinary(data)
	if err != nil {
		return fmt.Errorf("unable to unmarshal FromUrl in transaction, %v", err)
	}

	//get the length of the outputs
	toLen, data := common.BytesUint64(data[t.From.Size(nil):])

	if toLen > MaxTokenTxOutputs || toLen < 1 {
		return fmt.Errorf("invalid number of outputs for transaction")
	}
	t.To = make([]*TokenRecipient, toLen)
	i := 0
	for j := uint64(0); j < toLen; j++ {
		txOut := &TokenRecipient{}
		err := txOut.UnmarshalBinary(data[i:])
		if err != nil {
			return fmt.Errorf("unable to unmarshal token tx output for index %d", j)
		}
		i += txOut.Size()
		t.To[j] = txOut
	}

	if len(data) > i {
		//we have metadata
		b := types.Bytes{}
		err := b.UnmarshalBinary(data[i:])
		if err != nil {
			return fmt.Errorf("unable to unmarshal binary for meta data of transaction token tx")
		}
		t.Meta = b.Bytes()
	}

	return nil
}

// Size return the size of the Token Tx output
func (t *TokenRecipient) Size() int {
	var buf [16]byte
	count := binary.PutUvarint(buf[:], t.Amount)
	return t.URL.Size(nil) + count
}

// Equal returns true if t = t2
func (t *TokenRecipient) Equal(t2 *TokenRecipient) bool {
	defer func() { //                      ret will default to false, so if any error occurs
	}() //                                 we will return false as long as we catch any errors

	if t.Amount != t2.Amount {
		return false
	}
	if t.URL != t2.URL {
		return false
	}

	return true
}

// MarshalBinary serialize the token tx output
func (t *TokenRecipient) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer

	data, err := t.URL.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buffer.Write(data)

	data = common.Uint64Bytes(t.Amount)

	buffer.Write(data)
	return buffer.Bytes(), nil
}

// UnmarshalBinary deserialize the token tx output
func (t *TokenRecipient) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if rErr := recover(); rErr != nil {
			err = fmt.Errorf("error unmarshaling token tx output %v", rErr)
		}
	}()

	err = t.URL.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	l := t.URL.Size(nil)
	if l > len(data) {
		return fmt.Errorf("insufficient data to unmarshal amount")
	}
	t.Amount, _ = common.BytesUint64(data[l:])
	return nil
}
