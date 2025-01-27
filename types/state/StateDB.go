package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"sort"
	"sync"
	"time"

	"github.com/AccumulateNetwork/accumulate/internal/logging"
	"github.com/AccumulateNetwork/accumulate/smt/managed"
	"github.com/AccumulateNetwork/accumulate/smt/pmt"
	"github.com/AccumulateNetwork/accumulate/smt/storage"
	"github.com/AccumulateNetwork/accumulate/smt/storage/database"
	"github.com/AccumulateNetwork/accumulate/types"
	"github.com/tendermint/tendermint/libs/log"
)

type transactionStateInfo struct {
	Object  *Object
	ChainId types.Bytes
	TxId    types.Bytes
}

type transactionLists struct {
	validatedTx []*transactionStateInfo //list of validated transaction chain state objects for block
	pendingTx   []*transactionStateInfo //list of pending transaction chain state objects for block
	synthTxMap  map[types.Bytes32][]transactionStateInfo
}

// reset will (re)initialize the transaction lists, this should be done on startup and at the end of each block
func (t *transactionLists) reset() {
	t.pendingTx = nil
	t.validatedTx = nil
	t.synthTxMap = make(map[types.Bytes32][]transactionStateInfo)
}

type bucket string

const (
	bucketEntry            = bucket("StateEntries")
	bucketDataEntry        = bucket("DataEntries") //map of entry hash to WriteData Entry
	bucketTx               = bucket("Transactions")
	bucketMainToPending    = bucket("MainToPending") //main TXID to PendingTXID
	bucketPendingTx        = bucket("PendingTx")     //Store pending transaction
	bucketStagedSynthTx    = bucket("StagedSynthTx") //store the staged synthetic transactions
	bucketTxToSynthTx      = bucket("TxToSynthTx")   //TXID to synthetic TXID
	bucketMinorAnchorChain = bucket("MinorAnchorChain")

	markPower = int64(8)
)

//bucket SynthTx stores a list of synth tx's derived from a tx

func (b bucket) String() string { return string(b) }
func (b bucket) Bytes() []byte  { return []byte(b) }

type dataBlock struct {
	TxId      []byte //TxId is the reference to the transaction that created the data entry
	EntryHash []byte //EntryHash is the merkle root of the data entry
	Data      []byte //Data is the marshaled protocol.DataEntry
}

type dataBlockUpdates struct {
	bucket    bucket
	entries   []dataBlock
	stateData *Object //the data chain state
}

type blockUpdates struct {
	bucket    bucket
	txId      []*types.Bytes32
	stateData *Object //the latest chain state object modified from a tx
}

// StateDB the state DB will only retrieve information out of the database.  To store stuff use PersistentStateDB instead
type StateDB struct {
	dbMgr      *database.Manager
	merkleMgr  *managed.MerkleManager
	debug      bool
	bptMgr     *pmt.Manager //pbt is the global patricia trie for the application
	TimeBucket float64
	mutex      sync.Mutex
	sync       sync.WaitGroup
	logger     log.Logger
}

func (s *StateDB) logDebug(msg string, keyVals ...interface{}) {
	if s.logger != nil {
		s.logger.Debug(msg, keyVals...)
	}
}

func (s *StateDB) logInfo(msg string, keyVals ...interface{}) {
	if s.logger != nil {
		s.logger.Info(msg, keyVals...)
	}
}

func (s *StateDB) init(debug bool) {
	s.debug = debug

	s.bptMgr = pmt.NewBPTManager(s.dbMgr)
	managed.NewMerkleManager(s.dbMgr, markPower)
}

// Open database to manage the smt and chain states
func (s *StateDB) Open(dbFilename string, useMemDB bool, debug bool, logger log.Logger) (err error) {
	if logger != nil {
		s.logger = logger.With("module", "statedb")
	}

	var dbType, logName string
	if useMemDB {
		dbType, logName = "memory", "memdb"
	} else {
		dbType, logName = "badger", "badger"
	}

	if logger != nil {
		logger = logger.With("module", logName)
	}
	s.dbMgr, err = database.NewDBManager(dbType, dbFilename, logger)
	if err != nil {
		return err
	}

	s.merkleMgr, err = managed.NewMerkleManager(s.dbMgr, markPower)
	if err != nil {
		return err
	}

	s.init(debug)
	return nil
}

func (s *StateDB) Load(db storage.KeyValueDB, debug bool) (err error) {
	s.dbMgr = new(database.Manager)
	s.dbMgr.InitWithDB(db)
	s.merkleMgr, err = managed.NewMerkleManager(s.dbMgr, markPower)
	if err != nil {
		return err
	}

	s.init(debug)
	return nil
}

func (s *StateDB) GetDB() *database.Manager {
	return s.dbMgr
}

func (s *StateDB) Sync() {
	s.sync.Wait()
}

//GetChainRange get the transaction id's in a given range
func (s *StateDB) GetChainRange(chainId []byte, start int64, end int64) ([]types.Bytes32, int64, error) {
	mgr, err := s.ManageChain(chainId)
	if err != nil {
		return nil, 0, err
	}

	h, err := mgr.Entries(start, end)
	if err != nil {
		return nil, 0, err
	}

	resultHashes := make([]types.Bytes32, len(h))
	for i, h := range h {
		copy(resultHashes[i][:], h)
	}
	return resultHashes, mgr.Height(), nil
}

//GetChainDataByEntryHash retures the entry in the data chain by entry hash
func (s *StateDB) GetChainDataByEntryHash(chainId []byte, entryHash []byte) ([]byte, []byte, error) {
	data, err := s.dbMgr.Get(storage.MakeKey(bucketDataEntry, chainId, entryHash))
	if err != nil {
		return nil, nil, err
	}

	return entryHash, data, nil
}

//GetChainData retures the latest entry in the data chain
func (s *StateDB) GetChainData(chainId []byte) ([]byte, []byte, error) {
	mgr, err := s.ManageChain(bucketDataEntry, chainId)
	if err != nil {
		return nil, nil, err
	}

	currHeight := mgr.Height()
	entryHash, err := mgr.Entry(currHeight - 1)
	if err != nil {
		return nil, nil, err
	}

	return s.GetChainDataByEntryHash(chainId, entryHash)
}

//GetChainDataRange get the entryHashes in a given range
func (s *StateDB) GetChainDataRange(chainId []byte, start int64, end int64) ([]types.Bytes32, int64, error) {
	mgr, err := s.ManageChain(bucketDataEntry, chainId)
	if err != nil {
		return nil, 0, err
	}

	h, err := mgr.Entries(start, end)
	if err != nil {
		return nil, 0, err
	}

	resultHashes := make([]types.Bytes32, len(h))
	for i, h := range h {
		copy(resultHashes[i][:], h)
	}
	return resultHashes, mgr.Height(), nil
}

//GetTx get the transaction by transaction ID
func (s *StateDB) GetTx(txId []byte) (tx []byte, err error) {
	tx, err = s.dbMgr.Get(storage.MakeKey(bucketTx, txId))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

//GetPendingTx get the pending transactions by primary transaction ID
func (s *StateDB) GetPendingTx(txId []byte) (pendingTx []byte, err error) {

	pendingTxId, err := s.dbMgr.Get(storage.MakeKey(bucketMainToPending, txId))
	if err != nil {
		return nil, err
	}
	pendingTx, err = s.dbMgr.Get(storage.MakeKey(bucketPendingTx, pendingTxId))
	if err != nil {
		return nil, err
	}

	return pendingTx, nil
}

// GetSyntheticTxIds get the transaction id list by the transaction ID that spawned the synthetic transactions
func (s *StateDB) GetSyntheticTxIds(txId []byte) (syntheticTxIds []byte, err error) {

	syntheticTxIds, err = s.dbMgr.Get(storage.MakeKey(bucketTxToSynthTx, txId))
	if err != nil {
		//this is not a significant error. Synthetic transactions don't usually have other synth tx's.
		//TODO: Fixme, this isn't an error
		return nil, err
	}

	return syntheticTxIds, nil
}

// AddSynthTx add the synthetic transaction which is mapped to the parent transaction
func (tx *DBTransaction) AddSynthTx(parentTxId types.Bytes, synthTxId types.Bytes, synthTxObject *Object) {
	// TODO: Move the transaction related functions to stateTxn at a time where the impact on other branches is minimal

	tx.state.logDebug("AddSynthTx", "txid", logging.AsHex(synthTxId), "entry", logging.AsHex(synthTxObject.Entry))

	parentHash := parentTxId.AsBytes32()
	txMap := tx.transactions.synthTxMap
	txMap[parentHash] = append(txMap[parentHash], transactionStateInfo{synthTxObject, nil, synthTxId})
}

// AddDataEntry append the data entry to the data chain then update the
func (tx *DBTransaction) AddDataEntry(chainId *types.Bytes32, txid []byte, entryHash []byte, dataEntry []byte, dataState *Object) error {
	tx.state.logDebug("AddDataEntry", "chainId", logging.AsHex(chainId),
		"entryHash", logging.AsHex(entryHash), "entry", logging.AsHex(dataEntry))

	if len(entryHash) != 32 {
		return fmt.Errorf("cannot add data entry without an entry hash!")
	}

	if dataEntry == nil {
		return fmt.Errorf("cannot add data entry without an entry!")
	}

	tx.addDataEntry(chainId, txid, entryHash, dataEntry, dataState)

	return nil
}

// AddTransaction queues (pending) transaction signatures and (optionally) an
// accepted transaction for storage to their respective chains.
func (tx *DBTransaction) AddTransaction(chainId *types.Bytes32, txId types.Bytes, txPending, txAccepted *Object) error {
	var txAcceptedEntry []byte
	if txAccepted != nil {
		txAcceptedEntry = txAccepted.Entry
	}
	tx.state.logDebug("AddTransaction", "chainId", logging.AsHex(chainId), "txid", logging.AsHex(txId), "pending", logging.AsHex(txPending.Entry), "accepted", logging.AsHex(txAcceptedEntry))

	chainType, _ := binary.Uvarint(txPending.Entry)
	if types.ChainType(chainType) != types.ChainTypePendingTransaction {
		return fmt.Errorf("expecting pending transaction chain type of %s, but received %s",
			types.ChainTypePendingTransaction.Name(), types.ChainType(chainType).Name())
	}

	if txAccepted != nil {
		chainType, _ = binary.Uvarint(txAccepted.Entry)
		if types.ChainType(chainType) != types.ChainTypeTransaction {
			return fmt.Errorf("expecting pending transaction chain type of %s, but received %s",
				types.ChainTypeTransaction.Name(), types.ChainType(chainType).Name())
		}
	}

	//append the list of pending Tx's, txId's, and validated Tx's.
	tx.state.mutex.Lock()
	defer tx.state.mutex.Unlock()

	tsi := transactionStateInfo{txPending, chainId.Bytes(), txId}
	tx.transactions.pendingTx = append(tx.transactions.pendingTx, &tsi)

	if txAccepted != nil {
		tsi := transactionStateInfo{txAccepted, chainId.Bytes(), txId}
		tx.transactions.validatedTx = append(tx.transactions.validatedTx, &tsi)
	}
	return nil
}

//GetPersistentEntry will pull the data from the database for the StateEntries bucket.
func (s *StateDB) GetPersistentEntry(chainId []byte, verify bool) (*Object, error) {
	_ = verify
	s.Sync()

	if s.dbMgr == nil {
		return nil, fmt.Errorf("database has not been initialized")
	}

	data, err := s.dbMgr.Get(storage.MakeKey("StateEntries", chainId))
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%w: no state defined for %X", storage.ErrNotFound, chainId)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get state entry %X: %v", chainId, err)
	}

	ret := &Object{}
	err = ret.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state for %x", chainId)
	}
	//if verify {
	//todo: generate and verify data to make sure the state matches what is in the patricia trie
	//}
	return ret, nil
}

func (s *StateDB) getObject(keys ...interface{}) (*Object, error) {
	s.Sync()

	if s.dbMgr == nil {
		return nil, fmt.Errorf("database has not been initialized")
	}

	data, err := s.dbMgr.Get(storage.MakeKey(keys...))
	if err != nil {
		return nil, err
	}

	ret := &Object{}
	err = ret.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %v", err)
	}

	return ret, nil
}

// GetTransaction loads the state of the given transaction.
func (s *StateDB) GetTransaction(txid []byte) (*Object, error) {
	obj, err := s.getObject(bucketTx, txid)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%w: no transaction defined for %X", storage.ErrNotFound, txid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %X: %v", txid, err)
	}
	return obj, nil
}

// GetSynthTxn loads the state of the given staged synthetic transaction.
func (s *StateDB) GetSynthTxn(txid [32]byte) (*Object, error) {
	obj, err := s.getObject(bucketStagedSynthTx, "", txid)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%w: no transaction defined for %X", storage.ErrNotFound, txid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %X: %v", txid, err)
	}

	s.logDebug("GetSynthTxn", "txid", logging.AsHex(txid), "entry", logging.AsHex(obj.Entry))
	return obj, nil
}

// GetCurrentEntry retrieves the current state object from the database based upon chainId.  Current state either comes
// from a previously saves state for the current block, or it is from the database
func (tx *DBTransaction) GetCurrentEntry(chainId []byte) (*Object, error) {
	if chainId == nil {
		return nil, fmt.Errorf("chain id is invalid, thus unable to retrieve current entry")
	}
	var ret *Object
	var err error
	var key types.Bytes32

	copy(key[:32], chainId[:32])

	tx.state.mutex.Lock()
	currentState := tx.updates[key]
	tx.state.mutex.Unlock()
	if currentState != nil {
		ret = currentState.stateData
	} else {
		currentState := blockUpdates{}
		currentState.bucket = bucketEntry
		//pull current state entry from the database.
		currentState.stateData, err = tx.GetPersistentEntry(chainId, false)
		if err != nil {
			return nil, err
		}
		//if we have valid data, store off the state
		ret = currentState.stateData
	}

	return ret, nil
}

// AddStateEntry append the entry to the chain, the subChainId is if the chain upon which
// the transaction is against touches another chain. One example would be an account type chain
// may change the state of the KeyBook chain (i.e. a sub/secondary chain) based on the effect
// of a transaction.  The entry is the state object associated with
func (tx *DBTransaction) AddStateEntry(chainId *types.Bytes32, txHash *types.Bytes32, object *Object) {
	tx.state.logDebug("AddStateEntry", "chainId", logging.AsHex(chainId), "txHash", logging.AsHex(txHash), "entry", logging.AsHex(object.Entry))

	if txHash == nil {
		panic("Cannot add state entry without a transaction ID!")
	}

	tx.addStateEntry(chainId, txHash, object)
}

// UpdateNonce updates the state of a signator record _without adding to the
// transaction chain_.
func (tx *DBTransaction) UpdateNonce(chainId *types.Bytes32, object *Object) {
	tx.addStateEntry(chainId, nil, object)
}

func (tx *DBTransaction) addDataEntry(chainId *types.Bytes32, txid []byte, entryHash []byte, dataEntry []byte, dataState *Object) {
	tx.state.mutex.Lock()
	dataUpdates := tx.dataUpdates[*chainId]

	if dataUpdates == nil {
		dataUpdates = new(dataBlockUpdates)
		tx.dataUpdates[*chainId] = dataUpdates
	}

	if entryHash != nil {
		dataUpdates.entries = append(dataUpdates.entries, dataBlock{txid, entryHash[:], dataEntry})
	}

	dataUpdates.stateData = dataState
	tx.state.mutex.Unlock()

	//now update the chain state for the data chain
	txi := types.Bytes32{}
	txi.FromBytes(txid)
	tx.addStateEntry(chainId, &txi, dataState)
}

func (tx *DBTransaction) addStateEntry(chainId *types.Bytes32, txHash *types.Bytes32, object *Object) {
	begin := time.Now()

	tx.state.TimeBucket += float64(time.Since(begin)) * float64(time.Nanosecond) * 1e-9

	tx.state.mutex.Lock()
	updates := tx.updates[*chainId]
	tx.state.mutex.Unlock()

	if updates == nil {
		updates = new(blockUpdates)
		tx.updates[*chainId] = updates
	}

	if txHash != nil {
		updates.txId = append(updates.txId, txHash)
	}
	updates.stateData = object
}

func (tx *DBTransaction) WriteSynthTxn(txid []byte, obj *Object) error {
	data, err := obj.MarshalBinary()
	if err != nil {
		return err
	}

	var txid32 [32]byte
	copy(txid32[:], txid)

	tx.state.logInfo("Write synth txn", "txid", logging.AsHex(txid))

	tx.state.dbMgr.PutBatch(storage.MakeKey(bucketStagedSynthTx, "", txid), data)
	tx.state.bptMgr.Bpt.Insert(txid32, sha256.Sum256(data))
	return nil
}

func (tx *DBTransaction) writeTxs(mutex *sync.Mutex, group *sync.WaitGroup) error {
	defer group.Done()
	//record transactions
	for _, txn := range tx.transactions.validatedTx {
		data, _ := txn.Object.MarshalBinary()
		//store the transaction

		txHash := txn.TxId.AsBytes32()
		if synthTxInfos, ok := tx.transactions.synthTxMap[txHash]; ok {
			var synthData []byte
			for _, synthTxInfo := range synthTxInfos {
				synthData = append(synthData, synthTxInfo.TxId...)
				err := tx.WriteSynthTxn(synthTxInfo.TxId, synthTxInfo.Object)
				if err != nil {
					return err
				}
			}
			//store a list of txid to list of synth txid's
			tx.state.dbMgr.PutBatch(storage.MakeKey(bucketTxToSynthTx, txn.TxId), synthData)
		}

		mutex.Lock()
		//store the transaction in the transaction bucket by txid
		tx.state.dbMgr.PutBatch(storage.MakeKey(bucketTx, txn.TxId), data)
		//insert the hash of the tx object in the BPT
		tx.state.bptMgr.Bpt.Insert(txHash, sha256.Sum256(data))
		mutex.Unlock()
	}

	// record pending transactions
	for _, txn := range tx.transactions.pendingTx {
		//marshal the pending transaction state
		data, _ := txn.Object.MarshalBinary()
		//hash it and add to the merkle state for the pending chain
		pendingHash := sha256.Sum256(data)

		mutex.Lock()
		//Store the mapping of the Transaction hash to the pending transaction hash which can be used for
		// validation so we can find the pending transaction
		tx.state.dbMgr.PutBatch(storage.MakeKey("MainToPending", txn.TxId), pendingHash[:])

		//store the pending transaction by the pending tx hash
		tx.state.dbMgr.PutBatch(storage.MakeKey(bucketPendingTx, pendingHash[:]), data)
		mutex.Unlock()
	}

	return nil
}

func (tx *DBTransaction) writeChainState(group *sync.WaitGroup, mutex *sync.Mutex, mm *managed.MerkleManager, chainId types.Bytes32) error {
	defer group.Done()

	err := tx.state.merkleMgr.SetKey(storage.MakeKey(chainId[:]))
	if err != nil {
		return err
	}

	// We get ChainState objects here, instead. And THAT will hold
	//       the MerkleStateManager for the chain.
	//mutex.Lock()
	currentState := tx.updates[chainId]
	//mutex.Unlock()

	if currentState == nil {
		panic(fmt.Sprintf("Chain state is nil meaning no updates were stored on chain %X for the block. Should not get here!", chainId[:]))
	}

	//add all the transaction states that occurred during this block for this chain (in order of appearance)
	for _, txn := range currentState.txId {
		//store the txHash for the chains, they will be mapped back to the above recorded tx's
		tx.state.logDebug("AddHash", "hash", logging.AsHex(txn))
		mm.AddHash(managed.Hash((*txn)[:]))
	}

	if currentState.stateData != nil {
		//store the state of the main chain in the state object
		count := uint64(mm.MS.Count)
		currentState.stateData.Height = count
		currentState.stateData.Roots = make([][]byte, 64-bits.LeadingZeros64(count))
		for i := range currentState.stateData.Roots {
			if count&(1<<i) == 0 {
				// Only store the hashes we need
				continue
			}
			currentState.stateData.Roots[i] = mm.MS.Pending[i].Copy()
		}

		//now store the state object
		chainStateObject, err := currentState.stateData.MarshalBinary()
		if err != nil {
			panic("failed to marshal binary for state data")
		}

		mutex.Lock()
		tx.GetDB().PutBatch(storage.MakeKey(bucketEntry, chainId.Bytes()), chainStateObject)
		// The bpt stores the hash of the ChainState object hash.
		tx.state.bptMgr.Bpt.Insert(chainId, sha256.Sum256(chainStateObject))
		mutex.Unlock()
	}
	//TODO: figure out how to do this with new way state is derived
	//if len(currentState.pendingTx) != 0 {
	//	mdRoot := v.PendingChain.MS.GetMDRoot()
	//	if mdRoot == nil {
	//		//shouldn't get here, but will reject if I do
	//		panic(fmt.Sprintf("shouldn't get here on writeState() on chain id %X obtaining merkle state", chainId))
	//	}
	//	//todo:  Determine how we purge pending tx's after 2 weeks.
	//	s.bptMgr.Bpt.Insert(chainId, *mdRoot)
	//}

	return nil
}

func (tx *DBTransaction) writeDataState(chainId *types.Bytes32) error {
	mgr, err := tx.state.ManageChain(bucketDataEntry, chainId[:])
	if err != nil {
		return err
	}

	dataEntries := tx.dataUpdates[*chainId]

	if dataEntries == nil {
		return fmt.Errorf("not a data chain %x, no data was stored", chainId[:])
	}

	//add all the data entries that occurred during this block for this chain (in order of appearance)
	for _, entry := range dataEntries.entries {
		//store the entry hash for the data
		tx.state.logDebug("AddHash", "hash", logging.AsHex(entry.EntryHash))
		mgr.AddEntry(entry.EntryHash)
		tx.GetDB().PutBatch(storage.MakeKey(bucketDataEntry, chainId.Bytes(), entry.EntryHash), entry.Data)
	}

	// The bpt stores the root of the data merkle state
	anchor := types.Bytes32{}
	anchor.FromBytes(mgr.Anchor())
	tx.state.bptMgr.Bpt.Insert(storage.MakeKey(bucketDataEntry, chainId), anchor)
	return nil
}

func (s *StateDB) GetSynthTxnSigs() ([]SyntheticSignature, error) {
	chain, err := s.SynthTxidChain()
	if err != nil {
		return nil, err
	}

	record, err := chain.Record()
	switch {
	case err == nil:
		return record.Signatures, nil
	case errors.Is(err, storage.ErrNotFound):
		return nil, nil
	default:
		return nil, err
	}
}

func (tx *DBTransaction) writeSynthTxnSigs() error {
	sigMap := map[[32]byte]*SyntheticSignature{}
	delMap := map[[32]byte]bool{}

	// Remove any deleted entries
	for _, txid := range tx.delSynthSigs {
		delMap[txid] = true
	}

	// Load the chain
	chain, err := tx.state.SynthTxidChain()
	if err != nil {
		return err
	}

	// Get the current set of signatures
	var sigs []SyntheticSignature
	chainState, err := chain.Record()
	switch {
	case err == nil:
		sigs = chainState.Signatures
	case errors.Is(err, storage.ErrNotFound):
		// Ok
	default:
		return err
	}

	var changeCount int
	for _, sig := range sigs {
		if delMap[sig.Txid] {
			tx.state.logInfo("Removing synth txn sig", "txid", logging.AsHex(sig.Txid))
			changeCount++
			continue
		}

		tx.state.logInfo("Keeping synth txn sig", "txid", logging.AsHex(sig.Txid))
		sigMap[sig.Txid] = &sig
	}

	// Add new entries
	for _, sig := range tx.addSynthSigs {
		if sigMap[sig.Txid] != nil {
			continue
		}

		tx.state.logInfo("Adding synth txn sig", "txid", logging.AsHex(sig.Txid))
		changeCount++
		sigMap[sig.Txid] = sig
	}

	// If nothing changed, we're done
	if changeCount == 0 {
		return nil
	}

	// Create a sorted array of the transaction IDs
	txids := make([][32]byte, 0, len(sigMap))
	for txid := range sigMap {
		txids = append(txids, txid)
	}
	sort.Slice(txids, func(i, j int) bool {
		return bytes.Compare(txids[i][:], txids[j][:]) < 0
	})

	// Create a sorted array from the map
	sigs = make([]SyntheticSignature, 0, len(txids))
	for _, sig := range sigMap {
		sigs = append(sigs, *sig)
	}
	sort.Slice(sigs, func(i, j int) bool {
		return bytes.Compare(sigs[i].Txid[:], sigs[j].Txid[:]) < 0
	})

	// Update the chain state
	chainState.Signatures = sigs
	return chain.Chain.UpdateAs(chainState)
}

func (tx *DBTransaction) writeBatches() {
	defer tx.state.sync.Done()
	tx.state.dbMgr.EndBatch()
	tx.state.bptMgr.DBManager.EndBatch()
}

func (s *StateDB) SubnetID() (string, error) {
	b, err := s.GetDB().Get(storage.MakeKey("SubnetID"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *StateDB) BlockIndex() (int64, error) {
	mgr, err := s.MinorAnchorChain()
	if err != nil {
		return 0, err
	}

	if mgr.Height() == 0 {
		return 0, storage.ErrNotFound
	}

	head, err := mgr.Record()
	if err != nil {
		return 0, err
	}

	return head.Index, nil
}

// Commit will push the data to the database and update the patricia trie
func (tx *DBTransaction) Commit(blockHeight int64, timestamp time.Time, finalize func() error) ([]byte, error) {
	// //build a list of keys from the map
	// if !tx.dirty {
	// 	//only attempt to record the block if we have any data.
	// 	return tx.RootHash(), nil
	// }

	oldRoot := tx.RootHash()

	group := new(sync.WaitGroup)
	group.Add(1)
	group.Add(len(tx.updates))

	mutex := new(sync.Mutex)
	//to try the multi-threading add "go" in front of the next line
	err := tx.writeTxs(mutex, group)
	if err != nil {
		return nil, err
	}

	orderedUpdates := orderUpdateList(tx.updates)
	err = tx.commitUpdates(orderedUpdates, err, group, mutex)
	if err != nil {
		return nil, err
	}

	tx.commitTxWrites()

	err = tx.writeSynthTxnSigs()
	if err != nil {
		return nil, fmt.Errorf("failed to save synth txn signatures: %v", err)
	}

	err = tx.writeSynthChain(blockHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to write synth chain: %v", err)
	}

	tx.state.bptMgr.Bpt.Update()
	newRoot := tx.RootHash()

	group.Wait()

	// Update the anchor chain, but only if something changed
	//
	// Also update if the block height is 2 - this is a hack to make sure we
	// mirror properly, even if no records have changed
	if blockHeight == 2 || !bytes.Equal(oldRoot, newRoot) {
		err = tx.writeAnchorChain(blockHeight, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to write anchor chain: %v", err)
		}

		if finalize != nil {
			err = finalize()
			if err != nil {
				return nil, err
			}
		}

		tx.state.bptMgr.Bpt.Update()
	}

	//reset out block update buffer to get ready for the next round
	tx.state.sync.Add(1)
	//to enable threaded batch writes, put go in front of next line.
	tx.writeBatches()
	tx.cleanupTx()

	//return the state of the BPT for the state of the block
	rh := types.Bytes(tx.state.RootHash()).AsBytes32()
	tx.state.logDebug("WriteStates", "height", blockHeight, "hash", logging.AsHex(rh))
	return tx.state.RootHash(), nil
}

func orderUpdateList(updateList map[types.Bytes32]*blockUpdates) []types.Bytes32 {
	// Create an ordered list of chain IDs that need updating. The iteration
	// order of maps in Go is random. Randomly ordering database writes is bad,
	// because that leads to consensus errors between nodes, since each node
	// will have a different random order. So we need updates to have some
	// consistent order, regardless of what it is.
	updateOrder := make([]types.Bytes32, 0, len(updateList))
	for id := range updateList {
		updateOrder = append(updateOrder, id)
	}
	sort.Slice(updateOrder, func(i, j int) bool {
		return bytes.Compare(updateOrder[i][:], updateOrder[j][:]) < 0
	})
	return updateOrder
}

func (tx *DBTransaction) commitUpdates(orderedUpdates []types.Bytes32, err error, group *sync.WaitGroup, mutex *sync.Mutex) error {
	for _, chainId := range orderedUpdates {
		err = tx.writeChainState(group, mutex, tx.state.merkleMgr, chainId)
		if err != nil {
			return err
		}

		if _, ok := tx.dataUpdates[chainId]; ok {
			//if this is a data chain, then we'll have data entries we need to store
			err = tx.writeDataState(&chainId)
			if err != nil {
				return err
			}
		}

		//TODO: figure out how to do this with new way state is derived
		//if len(currentState.pendingTx) != 0 {
		//	mdRoot := v.PendingChain.MS.GetMDRoot()
		//	if mdRoot == nil {
		//		//shouldn't get here, but will reject if I do
		//		panic(fmt.Sprintf("shouldn't get here on writeState() on chain id %X obtaining merkle state", chainId))
		//	}
		//	//todo:  Determine how we purge pending tx's after 2 weeks.
		//	s.bptMgr.Bpt.Insert(chainId, *mdRoot)
		//}
	}
	return nil
}

func (tx *DBTransaction) commitTxWrites() {
	// Order the writes - can we skip this?
	writeOrder := make([]storage.Key, 0, len(tx.writes))
	for k := range tx.writes {
		writeOrder = append(writeOrder, k)
	}
	sort.Slice(writeOrder, func(i, j int) bool {
		return bytes.Compare(writeOrder[i][:], writeOrder[j][:]) < 0
	})
	for _, k := range writeOrder {
		tx.state.GetDB().PutBatch(storage.MakeKey(k), tx.writes[k])
	}
}

func (s *StateDB) RootHash() []byte {
	h := s.bptMgr.Bpt.Root.Hash // Make a copy
	return h[:]                 // Return a reference to the copy
}

func (tx *DBTransaction) cleanupTx() {
	tx.updates = make(map[types.Bytes32]*blockUpdates)
	tx.dataUpdates = make(map[types.Bytes32]*dataBlockUpdates)

	// The compiler optimizes this into a constant-time operation
	for k := range tx.writes {
		delete(tx.writes, k)
	}

	//clear out the transactions after they have been processed
	tx.transactions.validatedTx = nil
	tx.transactions.pendingTx = nil
	tx.transactions.synthTxMap = make(map[types.Bytes32][]transactionStateInfo)
}

func (s *StateDB) EnsureRootHash() []byte {
	s.bptMgr.Bpt.EnsureRootHash()
	return s.RootHash()
}
