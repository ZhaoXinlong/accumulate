package networks

import (
	"fmt"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

func MakeBouncer(networkList []int) *Bouncer {
	if len(networkList) > len(Networks) {
		return nil
	}

	rpcClients := []*rpchttp.HTTP{}
	for i := range networkList {
		lAddr := fmt.Sprintf("tcp://%s:%d", Networks[i].Ip[0], Networks[i].Port+1)
		client, err := rpchttp.New(lAddr, "/websocket")
		if err != nil {
			return nil
		}
		rpcClients = append(rpcClients, client)
	}
	txBouncer := NewBouncer(rpcClients)
	return txBouncer
}
