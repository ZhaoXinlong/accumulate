package chain

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/AccumulateNetwork/accumulate/config"
	"github.com/AccumulateNetwork/accumulate/internal/abci"
	"github.com/AccumulateNetwork/accumulate/internal/logging"
	"github.com/AccumulateNetwork/accumulate/internal/url"
	"github.com/AccumulateNetwork/accumulate/protocol"
	"github.com/AccumulateNetwork/accumulate/smt/common"
	"github.com/AccumulateNetwork/accumulate/smt/storage"
	"github.com/AccumulateNetwork/accumulate/types"
	"github.com/AccumulateNetwork/accumulate/types/api/transactions"
	"github.com/AccumulateNetwork/accumulate/types/state"
)

// synthCount returns the number of synthetic transactions sent by this subnet.
func (m *Executor) synthCount() (uint64, error) {
	k := storage.MakeKey("SyntheticTransactionCount")
	b, err := m.dbTx.Read(k)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return 0, err
	}

	var n uint64
	if len(b) > 0 {
		n, _ = common.BytesUint64(b)
	}
	return n, nil
}

// addSynthTxns prepares synthetic transactions for signing next block.
func (m *Executor) addSynthTxns(parentTxId types.Bytes, st *StateManager) error {
	// Need to pass this to a threaded batcher / dispatcher to do both signing
	// and sending of synth tx. No need to spend valuable time here doing that.
	for _, sub := range st.submissions {
		// Generate a synthetic tx and send to the router. Need to track txid to
		// make sure they get processed.

		tx, err := m.buildSynthTxn(sub.url, sub.body)
		if err != nil {
			return err
		}

		txSynthetic := state.NewPendingTransaction(tx)
		obj := new(state.Object)
		obj.Entry, err = txSynthetic.MarshalBinary()
		if err != nil {
			return err
		}

		m.dbTx.AddSynthTx(parentTxId, tx.TransactionHash(), obj)
	}

	return nil
}

func (m *Executor) addSystemTxns(txns ...*transactions.GenTransaction) error {
	if len(txns) == 0 {
		return nil
	}

	anchor, err := m.DB.MinorAnchorChain()
	if err != nil {
		return err
	}

	txids := make([][32]byte, len(txns))
	for i, tx := range txns {
		pending := state.NewPendingTransaction(tx)
		obj := new(state.Object)
		obj.Entry, err = pending.MarshalBinary()
		if err != nil {
			return err
		}

		err = m.dbTx.WriteSynthTxn(tx.TransactionHash(), obj)
		if err != nil {
			return err
		}

		copy(txids[i][:], tx.TransactionHash())
	}

	err = anchor.AddSystemTxns(txids...)
	if err != nil {
		return err
	}

	return nil
}

func (m *Executor) addAnchorTxn() error {
	synth, err := m.DB.SynthTxidChain()
	if err != nil {
		return err
	}

	synthHead, err := synth.Record()
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	anchor, err := m.DB.MinorAnchorChain()
	if err != nil {
		return err
	}

	anchorHead, err := anchor.Record()
	if err != nil {
		return err
	}

	switch {
	case anchorHead.Index == m.height && len(anchorHead.Chains) > 0:
		// Modified chains last block, continue
	case synthHead.Index == m.height:
		// Produced synthetic transactions last block, continue
	default:
		// Nothing happened last block, so skip creating an anchor txn
		return nil
	}

	body := new(protocol.SyntheticAnchor)
	body.Source = m.Network.NodeUrl().String()
	body.Index = m.height
	body.Timestamp = m.time
	copy(body.Root[:], m.DB.RootHash())
	body.Chains = anchorHead.Chains
	copy(body.ChainAnchor[:], anchor.Chain.Anchor())
	copy(body.SynthTxnAnchor[:], synth.Chain.Anchor())

	m.logDebug("Creating anchor txn", "root", logging.AsHex(body.Root), "chains", logging.AsHex(body.ChainAnchor), "synth", logging.AsHex(body.SynthTxnAnchor))

	var txns []*transactions.GenTransaction
	switch m.Network.Type {
	case config.Directory:
		// Send anchors from DN to all BVNs
		for _, bvn := range m.Network.BvnNames {
			tx, err := m.buildSynthTxn(protocol.BvnUrl(bvn), body)
			if err != nil {
				return err
			}
			txns = append(txns, tx)
		}

	case config.BlockValidator:
		// Send anchor from BVN to DN
		tx, err := m.buildSynthTxn(protocol.DnUrl(), body)
		if err != nil {
			return err
		}
		txns = append(txns, tx)
	}

	return m.addSystemTxns(txns...)
}

func (m *Executor) buildSynthTxn(dest *url.URL, body protocol.TransactionPayload) (*transactions.GenTransaction, error) {
	// Marshal the payload
	data, err := body.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal synthetic transaction payload: %v", err)
	}

	// Build the transaction
	tx := new(transactions.GenTransaction)
	tx.SigInfo = new(transactions.SignatureInfo)
	tx.SigInfo.URL = dest.String()
	tx.SigInfo.KeyPageHeight = 1
	tx.SigInfo.KeyPageIndex = 0
	tx.Transaction = data

	// Load the synth txn chain
	chain, err := m.DB.SynthTxidChain()
	if err != nil {
		return nil, err
	}

	// Load the chain state
	head, err := chain.Record()
	if errors.Is(err, storage.ErrNotFound) {
		head = new(state.SyntheticTransactionChain)
	} else if err != nil {
		return nil, err
	}

	// Increment the nonce
	head.Nonce++
	tx.SigInfo.Nonce = uint64(head.Nonce)

	// Save the updated chain state
	err = chain.Chain.UpdateAs(head)
	if err != nil {
		return nil, err
	}

	m.logDebug("Built synth txn", "txid", logging.AsHex(tx.TransactionHash()), "dest", dest.String(), "nonce", tx.SigInfo.Nonce, "type", body.GetType())

	return tx, nil
}

// signSynthTxns signs synthetic transactions from the previous block and
// prepares them to be sent next block.
func (m *Executor) signSynthTxns() error {
	// TODO If the leader fails, will the block happen again or will Tendermint
	// move to the next block? We need to be sure that synthetic transactions
	// won't get lost.

	// Load the synth txid chain
	chain, err := m.DB.SynthTxidChain()
	if err != nil {
		return err
	}

	// Retrieve transactions from the previous block
	txns, err := chain.LastBlock(m.height - 1)
	if err != nil {
		return err
	}

	// Check for anchor transactions from the previous block
	ac, err := m.DB.MinorAnchorChain()
	if err != nil {
		return err
	}

	acHead, err := ac.Record()
	switch {
	case err == nil:
		for _, txn := range acHead.SystemTxns {
			// Copy the variable. Otherwise we end up with a pointer to the loop
			// variable, which means we get multiple copies of the pointer, but
			// they all point to the same thing.
			txn := txn
			txns = append(txns, txn[:])
		}
	case errors.Is(err, storage.ErrNotFound):
		// Ok
	default:
		return fmt.Errorf("failed to retrieve anchor txn IDs")
	}

	// Only proceed if we have transactions to sign
	if len(txns) == 0 {
		return nil
	}

	// Use the synthetic transaction count to calculate what the nonces were
	nonce, err := m.synthCount()
	if err != nil {
		return err
	}

	// Sign all of the transactions
	body := new(protocol.SyntheticSignTransactions)
	for i, txid := range txns {
		m.logDebug("Signing synth txn", "txid", logging.AsHex(txid))

		// For each pending synthetic transaction
		var synthSig protocol.SyntheticSignature
		copy(synthSig.Txid[:], txid)

		// The nonce must be the final nonce minus (I + 1)
		synthSig.Nonce = nonce - 1 - uint64(i)

		// Sign it
		ed := new(transactions.ED25519Sig)
		ed.PublicKey = m.Key[32:]
		err = ed.Sign(synthSig.Nonce, m.Key, txid[:])
		if err != nil {
			return err
		}

		// Add it to the list
		synthSig.Signature = ed.Signature
		body.Transactions = append(body.Transactions, synthSig)
	}

	// Construct the signature transaction
	tx, err := m.buildSynthTxn(m.Network.NodeUrl(), body)
	if err != nil {
		return err
	}

	// Sign it
	ed := new(transactions.ED25519Sig)
	tx.Signature = append(tx.Signature, ed)
	ed.PublicKey = m.Key[32:]
	err = ed.Sign(tx.SigInfo.Nonce, m.Key, tx.TransactionHash())
	if err != nil {
		return err
	}

	// Marshal it
	data, err := tx.Marshal()
	if err != nil {
		return err
	}

	// Only the leader should actually send the transaction
	if !m.leader {
		return nil
	}

	// Send it
	go func() {
		_, err = m.Local.BroadcastTxAsync(context.Background(), data)
		if err != nil {
			m.logError("Failed to broadcast synth txn sigs", "error", err)
		}
	}()
	return nil
}

// sendSynthTxns sends signed synthetic transactions from previous blocks.
//
// Note, only the leader actually sends the transaction, but every other node
// must make the same updates to the database, otherwise consensus will fail,
// since constructing the synthetic transaction updates the nonce, which changes
// the BPT, so everyone needs to do that.
func (m *Executor) sendSynthTxns() ([]abci.SynthTxnReference, error) {
	// Get the signatures from the last block
	sigs, err := m.DB.GetSynthTxnSigs()
	if err != nil {
		return nil, err
	}

	// Is there anything to send?
	if len(sigs) == 0 {
		return nil, nil
	}

	// Array for synth TXN references
	refs := make([]abci.SynthTxnReference, 0, len(sigs))

	// Process all the transactions
	for _, sig := range sigs {
		// Load the pending transaction object
		obj, err := m.DB.GetSynthTxn(sig.Txid)
		if err != nil {
			return nil, err
		}

		// Unmarshal it
		state := new(state.PendingTransaction)
		err = obj.As(state)
		if err != nil {
			return nil, err
		}

		// Convert it back to a transaction
		tx := state.Restore()

		// Add the signature
		tx.Signature = append(tx.Signature, &transactions.ED25519Sig{
			Nonce:     sig.Nonce,
			PublicKey: sig.PublicKey,
			Signature: sig.Signature,
		})

		// Marshal the transaction
		raw, err := tx.Marshal()
		if err != nil {
			return nil, err
		}

		// Parse the URL
		u, err := url.Parse(tx.SigInfo.URL)
		if err != nil {
			return nil, err
		}

		// Add it to the batch
		m.logDebug("Sending synth txn", "actor", u.String(), "txid", logging.AsHex(tx.TransactionHash()))
		m.dispatcher.BroadcastTxAsync(context.Background(), u, raw)

		// Delete the signature
		m.dbTx.DeleteSynthTxnSig(sig.Txid)

		// Add the synthetic transaction reference
		var ref abci.SynthTxnReference
		ref.Type = uint64(tx.TransactionType())
		ref.Url = tx.SigInfo.URL
		ref.TxRef = sha256.Sum256(raw)
		copy(ref.Hash[:], tx.TransactionHash())
		refs = append(refs, ref)
	}

	return refs, nil
}
