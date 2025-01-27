package testing

import (
	"io"
	"time"

	"github.com/AccumulateNetwork/accumulate/config"
	"github.com/AccumulateNetwork/accumulate/internal/accumulated"
	"github.com/AccumulateNetwork/accumulate/internal/node"
	"github.com/AccumulateNetwork/accumulate/networks"
	tmnet "github.com/tendermint/tendermint/libs/net"
)

var LocalBVN = &networks.Subnet{
	Name: "Local",
	Type: config.BlockValidator,
	Port: 35550,
	Nodes: []networks.Node{
		{IP: "127.0.0.1", Type: config.Validator},
	},
}

func DefaultConfig(net config.NetworkType, node config.NodeType, netId string) *config.Config {
	cfg := config.Default(net, node, netId)        //
	cfg.Mempool.MaxBatchBytes = 1048576            //
	cfg.Mempool.CacheSize = 1048576                //
	cfg.Mempool.Size = 50000                       //
	cfg.Consensus.CreateEmptyBlocks = false        // Empty blocks are annoying to debug
	cfg.Consensus.TimeoutCommit = time.Second / 10 // Increase block frequency
	cfg.Accumulate.Website.Enabled = false         // No need for the website
	cfg.Instrumentation.Prometheus = false         // Disable prometheus: https://github.com/tendermint/tendermint/issues/7076
	cfg.Accumulate.Network.BvnNames = []string{netId}
	cfg.Accumulate.Network.Addresses = map[string][]string{netId: {"local"}}
	return cfg
}

func NodeInitOptsForNetwork(subnet *networks.Subnet) node.InitOptions {
	listenIP := make([]string, len(subnet.Nodes))
	remoteIP := make([]string, len(subnet.Nodes))
	cfg := make([]*config.Config, len(subnet.Nodes))

	for i, net := range subnet.Nodes {
		listenIP[i] = "localhost"
		remoteIP[i] = net.IP
		cfg[i] = DefaultConfig(subnet.Type, net.Type, subnet.Name) // Configure
	}

	port, err := tmnet.GetFreePort()
	if err != nil {
		panic(err)
	}

	return node.InitOptions{
		Port:     port,
		Config:   cfg,
		RemoteIP: remoteIP,
		ListenIP: listenIP,
	}
}

type DaemonOptions struct {
	Dir       string
	MemDB     bool
	LogWriter func(string) (io.Writer, error)
}

func RunDaemon(opts DaemonOptions, cleanup func(func())) (*accumulated.Daemon, error) {
	// Load the daemon
	daemon, err := accumulated.Load(opts.Dir, opts.LogWriter)
	if err != nil {
		return nil, err
	}

	// Set test knobs
	daemon.IsTest = true
	daemon.UseMemDB = opts.MemDB

	// Start the daemon
	err = daemon.Start()
	if err != nil {
		return nil, err
	}

	cleanup(func() {
		_ = daemon.Stop()
	})

	return daemon, nil
}
