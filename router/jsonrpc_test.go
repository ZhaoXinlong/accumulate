package router

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/AccumulateNetwork/accumulated/types"
	anon "github.com/AccumulateNetwork/accumulated/types/anonaddress"
	"github.com/AccumulateNetwork/accumulated/types/api"
	"github.com/AccumulateNetwork/accumulated/types/proto"
	"github.com/AccumulateNetwork/accumulated/types/synthetic"
	"github.com/go-playground/validator/v10"
	ptypes "github.com/tendermint/tendermint/abci/types"
)

//
//func sendFaucetTokenDeposit(client, address) {
//
//}

func TestJsonRpcAnonToken(t *testing.T) {

	_, privateKey, _ := ed25519.GenerateKey(nil)

	//make a client, and also spin up the router grpc
	dir, err := ioutil.TempDir("/tmp", "AccRouterTest-")
	cfg := path.Join(dir, "/config/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	client, _, _, rpcClient, vm := makeBVCandRouter(cfg, dir)

	if err != nil {
		t.Fatal(err)
	}

	query := NewQuery(vm)

	jsonapi := API{RandPort(), validator.New(), client, query}
	_ = jsonapi

	//create a key from the Tendermint node's private key. He will be the defacto source for the anon token.
	kpSponsor := ed25519.NewKeyFromSeed(vm.Key.PrivKey.Bytes()[:32])

	//use the public key of the bvc to make a sponsor address (this doesn't really matter right now, but need something so Identity of the BVC is good)
	adiSponsor := types.String(anon.GenerateAcmeAddress(kpSponsor.Public().(ed25519.PublicKey)))

	//set destination url address
	destAddress := types.String(anon.GenerateAcmeAddress(privateKey.Public().(ed25519.PublicKey)))

	txid := sha256.Sum256([]byte("fake txid"))

	tokenUrl := types.String("dc/ACME")

	//create a fake synthetic deposit for faucet.
	deposit := synthetic.NewTokenTransactionDeposit(txid[:], &adiSponsor, &destAddress)
	deposit.DepositAmount.SetInt64(500000000000)
	deposit.TokenUrl = tokenUrl

	depData, err := deposit.MarshalBinary()
	gtx := new(proto.GenTransaction)
	gtx.Transaction = depData
	if err := gtx.SetRoutingChainID(*destAddress.AsString()); err != nil {
		t.Fatal("bad url generated")
	}
	dataToSign := gtx.MarshalBinary()
	s := ed25519.Sign(privateKey, dataToSign)
	ed := new(proto.ED25519Sig)
	ed.PublicKey = privateKey[32:]
	ed.Signature = s
	gtx.Signature = append(gtx.Signature, ed)

	deliverRequestTXAsync := new(ptypes.RequestDeliverTx)
	deliverRequestTXAsync.Tx = gtx.Marshal()
	batch := rpcClient.NewBatch()
	batch.BroadcastTxAsync(context.Background(), deliverRequestTXAsync.Tx)

	Load(t, batch, privateKey)

	batch.Send(context.Background())

	//wait 3 seconds for the transaction to process for the block to complete.
	time.Sleep(3000 * time.Millisecond)
	queryTokenUrl := destAddress + "/" + tokenUrl
	resp, err := query.GetTokenAccount(queryTokenUrl.AsString())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(*resp.Data))
	output, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(output)

	//req := api.{}
	//adi := &api.ADI{}
	//adi.URL = "RoadRunner"
	//adi.PublicKeyHash = sha256.Sum256(privateKey.PubKey().Bytes())
	//data, err := json.Marshal(adi)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//req.Tx = &api.APIRequestRawTx{}
	//req.Tx.Signer = &api.Signer{}
	//req.Tx.Signer.URL = types.String(adiSponsor)
	//copy(req.Tx.Signer.PublicKey[:], kpSponsor.PubKey().Bytes())
	//req.Tx.Timestamp = time.Now().Unix()
	//adiJson := json.RawMessage(data)
	//req.Tx.Data = &adiJson
	//
	//ledger := types.MarshalBinaryLedgerAdiChainPath(*adi.URL.AsString(), *req.Tx.Data, req.Tx.Timestamp)
	//sig, err := kpSponsor.Sign(ledger)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//copy(req.Sig[:], sig)
	//
	//jsonReq, err := json.Marshal(&req)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	////now we can send in json rpc calls.
	//ret := jsonapi.faucet(context.Background(), jsonReq)

	//wait 30 seconds before shutting down.
	time.Sleep(30000 * time.Millisecond)

}

func Load(t *testing.T, rpcClient *rpchttp.BatchHTTP, Origin ed25519.PrivateKey) {
	srcURL := anon.GenerateAcmeAddress(Origin[32:])
	var SetOKeys []ed25519.PrivateKey
	var Addresses []string
	for i := 0; i < 1000; i++ {
		_, key, _ := ed25519.GenerateKey(nil)
		SetOKeys = append(SetOKeys, key)
		Addresses = append(Addresses, anon.GenerateAcmeAddress(key[32:]))
	}
	if len(Addresses) == 0 || len(SetOKeys) == 0 {
		t.Fatal("no addresses")
	}
	for i := 0; i < 1000; i++ {
		d := rand.Int() % len(SetOKeys)
		out := proto.Output{Dest: Addresses[d], Amount: 10 * 100000000}
		send := proto.NewTokenSend(srcURL, out)
		txData := send.Marshal()
		gtx := new(proto.GenTransaction)
		gtx.Transaction = txData
		//the send is routed to the srcUrl...
		if err := gtx.SetRoutingChainID(srcURL); err != nil {
			t.Fatal("bad url generated")
		}

		dataToSign := gtx.MarshalBinary()
		s := ed25519.Sign(Origin, dataToSign)

		ed := new(proto.ED25519Sig)
		ed.PublicKey = Origin[32:]
		ed.Signature = s
		gtx.Signature = append(gtx.Signature, ed)

		deliverRequestTXAsync := new(ptypes.RequestDeliverTx)
		deliverRequestTXAsync.Tx = gtx.Marshal()

		rpcClient.BroadcastTxAsync(context.Background(), deliverRequestTXAsync.Tx)
	}
}

func _TestJsonRpcAdi(t *testing.T) {

	//"wileecoyote/ACME"
	adiSponsor := "wileecoyote"

	kpNewAdi := types.CreateKeyPair()
	//routerAddress := fmt.Sprintf("tcp://localhost:%d", RandPort())

	//make a client, and also spin up the router grpc
	dir, err := ioutil.TempDir("/tmp", "AccRouterTest-")
	cfg := path.Join(dir, "/config/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	client, _, _, _, vm := makeBVCandRouter(cfg, dir)

	if err != nil {
		t.Fatal(err)
	}

	//kpSponsor := types.CreateKeyPair()

	query := NewQuery(vm)

	jsonapi := API{RandPort(), validator.New(), client, query}

	//StartAPI(RandPort(), client)

	kpSponsor := types.CreateKeyPairFromSeed(vm.Key.PrivKey.Bytes())

	req := api.APIRequestRaw{}
	adi := &api.ADI{}
	adi.URL = "RoadRunner"
	adi.PublicKeyHash = sha256.Sum256(kpNewAdi.PubKey().Bytes())
	data, err := json.Marshal(adi)
	if err != nil {
		t.Fatal(err)
	}

	req.Tx = &api.APIRequestRawTx{}
	req.Tx.Signer = &api.Signer{}
	req.Tx.Signer.URL = types.String(adiSponsor)
	copy(req.Tx.Signer.PublicKey[:], kpSponsor.PubKey().Bytes())
	req.Tx.Timestamp = time.Now().Unix()
	adiJson := json.RawMessage(data)
	req.Tx.Data = &adiJson

	ledger := types.MarshalBinaryLedgerAdiChainPath(*adi.URL.AsString(), *req.Tx.Data, req.Tx.Timestamp)
	sig, err := kpSponsor.Sign(ledger)
	if err != nil {
		t.Fatal(err)
	}
	copy(req.Sig[:], sig)

	jsonReq, err := json.Marshal(&req)
	if err != nil {
		t.Fatal(err)
	}

	//now we can send in json rpc calls.
	ret := jsonapi.createADI(context.Background(), jsonReq)

	t.Fatal(ret)

}
