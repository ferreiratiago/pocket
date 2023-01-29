package types

import (
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// DISCUSS_IN_THIS_COMMIT: Reevaluate the whole concept of a `TxResult`.

var _ modules.TxResult = &DefaultTxResult{}

func (txr *DefaultTxResult) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(txr)
}

func (*DefaultTxResult) FromBytes(bz []byte) (modules.TxResult, error) {
	result := new(DefaultTxResult)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (txr *DefaultTxResult) Hash() ([]byte, error) {
	bz, err := txr.Bytes()
	if err != nil {
		return nil, err
	}
	return txr.HashFromBytes(bz)
}

func (txr *DefaultTxResult) HashFromBytes(bz []byte) ([]byte, error) {
	return crypto.SHA3Hash(bz), nil
}

func (tx *Transaction) ToTxResult(height int64, index int, signer, recipient, msgType string, error Error) (*DefaultTxResult, Error) {
	txBytes, err := tx.Bytes()
	if err != nil {
		return nil, ErrProtoMarshal(err)
	}
	code, errString := int32(0), ""
	if error != nil {
		code = int32(error.Code())
		errString = err.Error()
	}
	return &DefaultTxResult{
		Tx:            txBytes,
		Height:        height,
		Index:         int32(index),
		ResultCode:    code,
		Error:         errString,
		SignerAddr:    signer,
		RecipientAddr: recipient,
		MessageType:   msgType,
	}, nil
}