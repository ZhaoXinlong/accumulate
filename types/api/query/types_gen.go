package query

// GENERATED BY go run ./internal/cmd/genmarshal. DO NOT EDIT.

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/AccumulateNetwork/accumulate/internal/encoding"
)

type RequestKeyPageIndex struct {
	Url string `json:"url,omitempty" form:"url" query:"url" validate:"required,acc-url"`
	Key []byte `json:"key,omitempty" form:"key" query:"key" validate:"required"`
}

type ResponseKeyPageIndex struct {
	KeyBook string `json:"keyBook,omitempty" form:"keyBook" query:"keyBook" validate:"required"`
	KeyPage string `json:"keyPage,omitempty" form:"keyPage" query:"keyPage" validate:"required"`
	Index   uint64 `json:"index" form:"index" query:"index" validate:"required"`
}

func (v *RequestKeyPageIndex) Equal(u *RequestKeyPageIndex) bool {
	if !(v.Url == u.Url) {
		return false
	}

	if !(bytes.Equal(v.Key, u.Key)) {
		return false
	}

	return true
}

func (v *ResponseKeyPageIndex) Equal(u *ResponseKeyPageIndex) bool {
	if !(v.KeyBook == u.KeyBook) {
		return false
	}

	if !(v.KeyPage == u.KeyPage) {
		return false
	}

	if !(v.Index == u.Index) {
		return false
	}

	return true
}

func (v *RequestKeyPageIndex) BinarySize() int {
	var n int

	n += encoding.StringBinarySize(v.Url)

	n += encoding.BytesBinarySize(v.Key)

	return n
}

func (v *ResponseKeyPageIndex) BinarySize() int {
	var n int

	n += encoding.StringBinarySize(v.KeyBook)

	n += encoding.StringBinarySize(v.KeyPage)

	n += encoding.UvarintBinarySize(v.Index)

	return n
}

func (v *RequestKeyPageIndex) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(encoding.StringMarshalBinary(v.Url))

	buffer.Write(encoding.BytesMarshalBinary(v.Key))

	return buffer.Bytes(), nil
}

func (v *ResponseKeyPageIndex) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(encoding.StringMarshalBinary(v.KeyBook))

	buffer.Write(encoding.StringMarshalBinary(v.KeyPage))

	buffer.Write(encoding.UvarintMarshalBinary(v.Index))

	return buffer.Bytes(), nil
}

func (v *RequestKeyPageIndex) UnmarshalBinary(data []byte) error {
	if x, err := encoding.StringUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding Url: %w", err)
	} else {
		v.Url = x
	}
	data = data[encoding.StringBinarySize(v.Url):]

	if x, err := encoding.BytesUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding Key: %w", err)
	} else {
		v.Key = x
	}
	data = data[encoding.BytesBinarySize(v.Key):]

	return nil
}

func (v *ResponseKeyPageIndex) UnmarshalBinary(data []byte) error {
	if x, err := encoding.StringUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding KeyBook: %w", err)
	} else {
		v.KeyBook = x
	}
	data = data[encoding.StringBinarySize(v.KeyBook):]

	if x, err := encoding.StringUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding KeyPage: %w", err)
	} else {
		v.KeyPage = x
	}
	data = data[encoding.StringBinarySize(v.KeyPage):]

	if x, err := encoding.UvarintUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding Index: %w", err)
	} else {
		v.Index = x
	}
	data = data[encoding.UvarintBinarySize(v.Index):]

	return nil
}

func (v *RequestKeyPageIndex) MarshalJSON() ([]byte, error) {
	u := struct {
		Url string  `json:"url,omitempty"`
		Key *string `json:"key,omitempty"`
	}{}
	u.Url = v.Url
	u.Key = encoding.BytesToJSON(v.Key)
	return json.Marshal(&u)
}

func (v *RequestKeyPageIndex) UnmarshalJSON(data []byte) error {
	u := struct {
		Url string  `json:"url,omitempty"`
		Key *string `json:"key,omitempty"`
	}{}
	u.Url = v.Url
	u.Key = encoding.BytesToJSON(v.Key)
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Url = u.Url
	if x, err := encoding.BytesFromJSON(u.Key); err != nil {
		return fmt.Errorf("error decoding Key: %w", err)
	} else {
		v.Key = x
	}
	return nil
}
