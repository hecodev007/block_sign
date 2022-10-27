// Contains the node database, storing previously seen nodes and any collected
// metadata about them for QoS purposes.

package dht

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"os"
	"path"
	"sync"
	"time"

	wire "btmSign/bytom/lib/github.com/tendermint/go-wire"
	log "github.com/sirupsen/logrus"

	"btmSign/bytom/crypto"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/errors"
)

var (
	nodeDBNilNodeID      = NodeID{}       // Special node ID to use as a nil element.
	nodeDBNodeExpiration = 24 * time.Hour // Time after which an unseen node should be dropped.
	nodeDBCleanupCycle   = time.Hour      // Time period for running the expiration task.
)

// nodeDB stores all nodes we know about.
type nodeDB struct {
	lvl    dbm.DB        // Interface to the database itself
	self   NodeID        // Own node id to prevent adding it into the database
	runner sync.Once     // Ensures we can start at most one expirer
	quit   chan struct{} // Channel to signal the expiring thread to stop
}

// Schema layout for the node database
var (
	nodeDBVersionKey = []byte("version") // Version of the database to flush if changes
	nodeDBItemPrefix = []byte("n:")      // Identifier to prefix node entries with

	nodeDBDiscoverRoot      = ":discover"
	nodeDBDiscoverPing      = nodeDBDiscoverRoot + ":lastping"
	nodeDBDiscoverPong      = nodeDBDiscoverRoot + ":lastpong"
	nodeDBDiscoverFindFails = nodeDBDiscoverRoot + ":findfail"
	nodeDBTopicRegTickets   = ":tickets"
)

// newNodeDB creates a new node database for storing and retrieving infos about
// known peers in the network. If no path is given, an in-memory, temporary
// database is constructed.
func newNodeDB(path string, version int, self NodeID) (*nodeDB, error) {
	if path == "" {
		return newMemoryNodeDB(self), nil
	}
	return newPersistentNodeDB(path, version, self)
}

// newMemoryNodeDB creates a new in-memory node database without a persistent
// backend.
func newMemoryNodeDB(self NodeID) *nodeDB {
	db := dbm.NewMemDB()
	return &nodeDB{
		lvl:  db,
		self: self,
		quit: make(chan struct{}),
	}
}

// newPersistentNodeDB creates/opens a leveldb backed persistent node database,
// also flushing its contents in case of a version mismatch.
func newPersistentNodeDB(filePath string, version int, self NodeID) (*nodeDB, error) {
	dir, file := path.Split(filePath)
	if file == "" {
		return nil, errors.New("unspecified db file name")
	}
	db := dbm.NewDB(file, dbm.GoLevelDBBackendStr, dir)

	// The nodes contained in the cache correspond to a certain protocol version.
	// Flush all nodes if the version doesn't match.
	currentVer := make([]byte, binary.MaxVarintLen64)
	currentVer = currentVer[:binary.PutVarint(currentVer, int64(version))]

	blob := db.Get(nodeDBVersionKey)
	if blob == nil {
		db.Set(nodeDBVersionKey, currentVer)
	} else if !bytes.Equal(blob, currentVer) {
		db.Close()
		if err := os.RemoveAll(filePath + ".db"); err != nil {
			return nil, err
		}
		return newPersistentNodeDB(filePath, version, self)
	}

	return &nodeDB{
		lvl:  db,
		self: self,
		quit: make(chan struct{}),
	}, nil
}

// makeKey generates the leveldb key-blob from a node id and its particular
// field of interest.
func makeKey(id NodeID, field string) []byte {
	if bytes.Equal(id[:], nodeDBNilNodeID[:]) {
		return []byte(field)
	}
	return append(nodeDBItemPrefix, append(id[:], field...)...)
}

// splitKey tries to split a database key into a node id and a field part.
func splitKey(key []byte) (id NodeID, field string) {
	// If the key is not of a node, return it plainly
	if !bytes.HasPrefix(key, nodeDBItemPrefix) {
		return NodeID{}, string(key)
	}
	// Otherwise split the id and field
	item := key[len(nodeDBItemPrefix):]
	copy(id[:], item[:len(id)])
	field = string(item[len(id):])

	return id, field
}

// fetchInt64 retrieves an integer instance associated with a particular
// database key.
func (db *nodeDB) fetchInt64(key []byte) int64 {
	blob := db.lvl.Get(key)
	if blob == nil {
		return 0
	}
	val, read := binary.Varint(blob)
	if read <= 0 {
		return 0
	}
	return val
}

// storeInt64 update a specific database entry to the current time instance as a
// unix timestamp.
func (db *nodeDB) storeInt64(key []byte, n int64) {
	blob := make([]byte, binary.MaxVarintLen64)
	blob = blob[:binary.PutVarint(blob, n)]
	db.lvl.Set(key, blob)
}

// node retrieves a node with a given id from the database.
func (db *nodeDB) node(id NodeID) *Node {
	var (
		n    = int(0)
		node = new(Node)
	)

	key := makeKey(id, nodeDBDiscoverRoot)
	rawData := db.lvl.Get(key)

	var err error
	wire.ReadBinary(node, bytes.NewReader(rawData), 0, &n, &err)
	if err != nil {
		log.WithFields(log.Fields{"module": logModule, "key": key, "node": node, "error": err}).Warn("get node from db err")
		return nil
	}

	node.sha = crypto.Sha256Hash(node.ID[:])
	return node
}

// updateNode inserts - potentially overwriting - a node into the peer database.
func (db *nodeDB) updateNode(node *Node) error {
	var (
		n    = int(0)
		err  = error(nil)
		blob = new(bytes.Buffer)
	)

	wire.WriteBinary(node, blob, &n, &err)
	if err != nil {
		return err
	}

	db.lvl.Set(makeKey(node.ID, nodeDBDiscoverRoot), blob.Bytes())
	return nil
}

// deleteNode deletes all information/keys associated with a node.
func (db *nodeDB) deleteNode(id NodeID) {
	deleter := db.lvl.IteratorPrefix(makeKey(id, ""))
	for deleter.Next() {
		db.lvl.Delete(deleter.Key())
	}
}

// ensureExpirer is a small helper method ensuring that the data expiration
// mechanism is running. If the expiration goroutine is already running, this
// method simply returns.
//
// The goal is to start the data evacuation only after the network successfully
// bootstrapped itself (to prevent dumping potentially useful seed nodes). Since
// it would require significant overhead to exactly trace the first successful
// convergence, it's simpler to "ensure" the correct state when an appropriate
// condition occurs (i.e. a successful bonding), and discard further events.
func (db *nodeDB) ensureExpirer() {
	db.runner.Do(func() { go db.expirer() })
}

// expirer should be started in a go routine, and is responsible for looping ad
// infinitum and dropping stale data from the database.
func (db *nodeDB) expirer() {
	tick := time.NewTicker(nodeDBCleanupCycle)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if err := db.expireNodes(); err != nil {
				log.WithFields(log.Fields{"module": logModule, "error": err}).Error("Failed to expire nodedb items")
			}
		case <-db.quit:
			return
		}
	}
}

// expireNodes iterates over the database and deletes all nodes that have not
// been seen (i.e. received a pong from) for some allotted time.
func (db *nodeDB) expireNodes() error {
	threshold := time.Now().Add(-nodeDBNodeExpiration)

	// Find discovered nodes that are older than the allowance
	it := db.lvl.Iterator()
	defer it.Release()

	for it.Next() {
		// Skip the item if not a discovery node
		id, field := splitKey(it.Key())
		if field != nodeDBDiscoverRoot {
			continue
		}
		// Skip the node if not expired yet (and not self)
		if !bytes.Equal(id[:], db.self[:]) {
			if seen := db.lastPong(id); seen.After(threshold) {
				continue
			}
		}
		// Otherwise delete all associated information
		db.deleteNode(id)
	}

	return nil
}

// lastPing retrieves the time of the last ping packet send to a remote node,
// requesting binding.
func (db *nodeDB) lastPing(id NodeID) time.Time {
	return time.Unix(db.fetchInt64(makeKey(id, nodeDBDiscoverPing)), 0)
}

// updateLastPing updates the last time we tried contacting a remote node.
func (db *nodeDB) updateLastPing(id NodeID, instance time.Time) {
	db.storeInt64(makeKey(id, nodeDBDiscoverPing), instance.Unix())
}

// lastPong retrieves the time of the last successful contact from remote node.
func (db *nodeDB) lastPong(id NodeID) time.Time {
	return time.Unix(db.fetchInt64(makeKey(id, nodeDBDiscoverPong)), 0)
}

// updateLastPong updates the last time a remote node successfully contacted.
func (db *nodeDB) updateLastPong(id NodeID, instance time.Time) {
	db.storeInt64(makeKey(id, nodeDBDiscoverPong), instance.Unix())
}

// findFails retrieves the number of findnode failures since bonding.
func (db *nodeDB) findFails(id NodeID) int {
	return int(db.fetchInt64(makeKey(id, nodeDBDiscoverFindFails)))
}

// updateFindFails updates the number of findnode failures since bonding.
func (db *nodeDB) updateFindFails(id NodeID, fails int) {
	db.storeInt64(makeKey(id, nodeDBDiscoverFindFails), int64(fails))
}

// querySeeds retrieves random nodes to be used as potential seed nodes
// for bootstrapping.
func (db *nodeDB) querySeeds(n int, maxAge time.Duration) []*Node {
	var (
		now   = time.Now()
		nodes = make([]*Node, 0, n)
		it    = db.lvl.Iterator()
		id    NodeID
	)
	defer it.Release()

seek:
	for seeks := 0; len(nodes) < n && seeks < n*5; seeks++ {
		// Seek to a random entry. The first byte is incremented by a
		// random amount each time in order to increase the likelihood
		// of hitting all existing nodes in very small databases.
		ctr := id[0]
		if _, err := rand.Read(id[:]); err != nil {
			log.WithFields(log.Fields{"module": logModule, "error": err}).Warn("get rand date")
		}
		id[0] = ctr + id[0]%16
		it.Seek(makeKey(id, nodeDBDiscoverRoot))

		n := nextNode(it)
		if n == nil {
			id[0] = 0
			continue seek // iterator exhausted
		}
		if n.ID == db.self {
			continue seek
		}
		if now.Sub(db.lastPong(n.ID)) > maxAge {
			continue seek
		}
		for i := range nodes {
			if nodes[i].ID == n.ID {
				continue seek // duplicate
			}
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func (db *nodeDB) fetchTopicRegTickets(id NodeID) (issued, used uint32) {
	key := makeKey(id, nodeDBTopicRegTickets)
	blob := db.lvl.Get(key)

	if len(blob) != 8 {
		return 0, 0
	}
	issued = binary.BigEndian.Uint32(blob[0:4])
	used = binary.BigEndian.Uint32(blob[4:8])
	return
}

func (db *nodeDB) updateTopicRegTickets(id NodeID, issued, used uint32) {
	key := makeKey(id, nodeDBTopicRegTickets)
	blob := make([]byte, 8)
	binary.BigEndian.PutUint32(blob[0:4], issued)
	binary.BigEndian.PutUint32(blob[4:8], used)
	db.lvl.Set(key, blob)
}

// reads the next node record from the iterator, skipping over other
// database entries.
func nextNode(it dbm.Iterator) *Node {
	var (
		n    = int(0)
		err  = error(nil)
		node = new(Node)
	)

	for end := false; !end; end = !it.Next() {
		id, field := splitKey(it.Key())
		if field != nodeDBDiscoverRoot {
			continue
		}

		wire.ReadBinary(node, bytes.NewReader(it.Value()), 0, &n, &err)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "id": id, "error": err}).Error("invalid node")
			continue
		}

		return node
	}
	return nil
}

// close flushes and closes the database files.
func (db *nodeDB) close() {
	close(db.quit)
	db.lvl.Close()
}
