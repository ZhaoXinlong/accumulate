package validator

import (
	pb "github.com/AccumulateNetwork/accumulated/types/proto"

	//"crypto/sha256"
	"fmt"
	//"github.com/AccumulateNetwork/SMT/managed"
	acctypes "github.com/AccumulateNetwork/accumulated/types/state"
	cfg "github.com/tendermint/tendermint/config"
	//dbm "github.com/tendermint/tm-db"
	"time"
)

type CreateIdentityValidator struct {
	ValidatorContext

	EV *EntryValidator
}

//transactions are just accounts with balances on a given token chain
//what transaction types should be supported?
//type Identity struct {
//	Version int8
//    DDII string
//    PublicKey managed.Hash
//}
//
//func (tx *Identity) MarshalBinary() ([]byte, error){
//	ret := make([]byte, )
//	return nil, nil
//}
//
//func (tx *Identity) UnmarshalBinary(data []byte) error{
//
//	return nil
//}

func NewCreateIdentityValidator() *CreateIdentityValidator {
	v := CreateIdentityValidator{}
	//need the chainid, then hash to get first 8 bytes to make the chainid.
	//by definition a chainid of a factoid block is
	//000000000000000000000000000000000000000000000000000000000000000f
	//the id will be 0x0000000f
	chainid := "000000000000000000000000000000000000000000000000000000000000001D" //does this make sense anymore?
	v.EV = NewEntryValidator()
	v.SetInfo(chainid, "create-identity", pb.AccInstruction_Identity_Creation)
	v.ValidatorContext.ValidatorInterface = &v
	return &v
}

func (v *CreateIdentityValidator) Check(currentstate *StateEntry, identitychain []byte, chainid []byte, p1 uint64, p2 uint64, data []byte) error {
	return nil
}
func (v *CreateIdentityValidator) Initialize(config *cfg.Config) error {
	return nil
}

func (v *CreateIdentityValidator) BeginBlock(height int64, time *time.Time) error {
	v.lastHeight = v.currentHeight
	v.lastTime = v.currentTime
	v.currentHeight = height
	v.currentTime = *time

	return nil
}

func (v *CreateIdentityValidator) Validate(currentstate *StateEntry, submission *pb.Submission) (resp *ResponseValidateTX, err error) {
	if currentstate == nil {
		//but this is to be expected...
		return nil, fmt.Errorf("Current State Not Defined")
	}

	//Temporary validation rules:
	idstate := acctypes.IdentityState{}
	err = idstate.UnmarshalBinary(submission.Data)
	if err != nil {
		return nil, err
	}

	resp = &ResponseValidateTX{}
	//so. also need to return the identity chain and chain id these belong to....  Really need the factom entry format updated.
	resp.StateData = submission.Data //make([][]byte,1)
	//resp.StateData[0] = data

	return resp, nil
	//this builds the entry if valid
	_, err = v.EV.Validate(currentstate, submission)

	if err != nil {
		return nil, err
	}

	//now we need to validate the contents.
	//for _ := range res.Submissions {
	//	//now we need to validate the contents.
	//	//need to validate this: res.Submissions[i].Data()

	return nil, nil
	//return &pb.Submission{}, nil
}

func (v *CreateIdentityValidator) EndBlock(mdroot []byte) error {
	//copy(v.mdroot[:], mdroot[:])
	//don't think this serves a purpose???
	return nil
}
