package skademlia

import (
	"container/heap"
	"sync"
	"time"

	"github.com/guyu96/noise"
	"github.com/guyu96/noise/protocol"
)

// Broadcast sends a message denoted by its opcode and content to all S/Kademlia IDs
// closest in terms of XOR distance to that of a specified node instances ID.
//
// Every message sent will be blocking. To have every message sent not block the current
// goroutine, refer to `BroadcastAsync(node *noise.Node, message noise.Message)`
//
// It returns a list of errors which have occurred in sending any messages to peers
// closest to a given node instance.
func Broadcast(node *noise.Node, message noise.Message) (errs []error) {
	errorChannels := make([]<-chan error, 0)

	for _, peerID := range FindClosestPeers(Table(node), protocol.NodeID(node).Hash(), BucketSize()) {
		peer := protocol.Peer(node, peerID)

		if peer == nil {
			continue
		}

		errorChannels = append(errorChannels, peer.SendMessageAsync(message))
	}

	for _, ch := range errorChannels {
		err := <-ch
		if err != nil {
			errs = append(errs, err)
		}
	}

	return
}

// BroadcastAsync sends a message denoted by its opcode and content to all S/Kademlia IDs
// closest in terms of XOR distance to that of a specified node instances ID.
//
// Every message sent will be non-blocking. To have every message sent block the current
// goroutine, refer to `Broadcast(node *noise.Node, message noise.Message) (errs []error)`
func BroadcastAsync(node *noise.Node, message noise.Message) {
	for _, peerID := range FindClosestPeers(Table(node), protocol.NodeID(node).Hash(), BucketSize()) {
		peer := protocol.Peer(node, peerID)

		if peer == nil {
			continue
		}

		peer.SendMessageAsync(message)
	}
}

func queryPeerByID(node *noise.Node, peerID, targetID ID, responses chan []ID) {
	var err error

	if peerID.Equals(protocol.NodeID(node)) {
		responses <- []ID{}
		return
	}

	peer := protocol.Peer(node, peerID)

	if peer == nil {
		peer, err = node.Dial(peerID.address)

		if err != nil {
			responses <- []ID{}
			return
		}

		WaitUntilAuthenticated(peer)
	}

	opcodeLookupResponse, err := noise.OpcodeFromMessage((*LookupResponse)(nil))
	if err != nil {
		panic("skademlia: response opcode not registered")
	}

	// Send lookup request.
	err = peer.SendMessage(LookupRequest{ID: targetID})
	if err != nil {
		responses <- []ID{}
		return
	}

	// Handle lookup response.
	select {
	case msg := <-peer.Receive(opcodeLookupResponse):
		responses <- msg.(LookupResponse).peers
	case <-time.After(3 * time.Second):
		responses <- []ID{}
	}
}

type lookupBucket struct {
	pending int
	queue   []ID
}

func (lookup *lookupBucket) performLookup(node *noise.Node, table *table, targetID ID, alpha int, visited *sync.Map) (results []ID) {
	responses := make(chan []ID)

	// Go through every peer in the entire queue and queue up what peers believe
	// is closest to a target ID.
	for ; lookup.pending < alpha && len(lookup.queue) > 0; lookup.pending++ {
		go queryPeerByID(node, lookup.queue[0], targetID, responses)
		lookup.queue = lookup.queue[1:]
	}

	// Empty queue.
	lookup.queue = lookup.queue[:0]

	// Asynchronous breadth-first search.
	for lookup.pending > 0 {
		response := <-responses

		lookup.pending--

		// Expand responses containing a peer's belief on the closest peers to target ID.
		for _, id := range response {
			if _, seen := visited.LoadOrStore(string(id.Hash()), struct{}{}); !seen {
				// Append new peer to be queued by the routing table.
				results = append(results, id)
				lookup.queue = append(lookup.queue, id)
			}
		}

		// Queue and request for #ALPHA closest peers to target ID from expanded results.
		for ; lookup.pending < alpha && len(lookup.queue) > 0; lookup.pending++ {
			go queryPeerByID(node, lookup.queue[0], targetID, responses)
			lookup.queue = lookup.queue[1:]
		}

		// Empty queue.
		lookup.queue = lookup.queue[:0]
	}

	return
}

func foundNode(ids []protocol.ID, target ID) int {
	for i, id := range ids {
		if id.Equals(target) {
			return i
		}
	}
	return -1
}

// FindNode implements the `FIND_NODE` RPC method denoted in
// Section 4.4 of the S/Kademlia paper: `Lookup over disjoint paths`.
//
// Given a node instance N, and a S/Kademlia ID as a target T, α disjoint lookups
// take place in parallel over all closest peers to N to target T, with at most D
// lookups happening at once.
//
// Each disjoint lookup queries at most α peers.
//
// It returns at most BUCKET_SIZE S/Kademlia peer IDs closest to that of a
// specified target T.
func FindNode(node *noise.Node, targetID ID, alpha int, numDisjointPaths int) []ID {
	table, visited := Table(node), new(sync.Map)

	visited.Store(string(protocol.NodeID(node).Hash()), struct{}{})
	visited.Store(string(targetID.Hash()), struct{}{})

	var lookups []*lookupBucket

	results := PriorityQueue{}
	heap.Init(&results)

	// Start searching for target from α peers closest to T by queuing
	// them up and marking them as visited.
	targetHash := targetID.Hash()
	localClosest := FindClosestPeers(table, targetHash, alpha)
	// if foundNode(localClosest, targetID) != -1 {
	// 	return []ID{targetID}
	// }
	for i, peerID := range localClosest {
		visited.Store(string(peerID.Hash()), struct{}{})

		if len(lookups) < numDisjointPaths {
			lookups = append(lookups, new(lookupBucket))
		}

		lookup := lookups[i%numDisjointPaths]
		lookup.queue = append(lookup.queue, peerID.(ID))

		item := &Item{
			value:    peerID.(ID),
			priority: xorPriority(targetHash, peerID.(ID).Hash()),
		}
		heap.Push(&results, item)
	}

	var wait sync.WaitGroup
	var mutex sync.Mutex

	for _, lookup := range lookups {
		go func(lookup *lookupBucket) {
			mutex.Lock()
			res := lookup.performLookup(node, table, targetID, alpha, visited)

			for _, peerID := range res {
				item := &Item{
					value:    peerID,
					priority: xorPriority(targetHash, peerID.Hash()),
				}
				heap.Push(&results, item)
			}
			if len(results) > alpha {
				results = results[:alpha] // only keep at most alpha nubmer of closest peers
			}
			mutex.Unlock()

			wait.Done()
		}(lookup)

		wait.Add(1)
	}

	// wait until all D parallel lookups have been completed.
	wait.Wait()

	ids := []ID{}
	for results.Len() > 0 {
		item := heap.Pop(&results).(*Item)
		ids = append(ids, item.value.(ID))
	}

	// reverse ids, which is ordered in in ascending distance
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}

	return ids
}

// func FindNode(node *noise.Node, targetID ID, alpha int, numDisjointPaths int) (results []ID) {
// 	table, visited := Table(node), new(sync.Map)

// 	visited.Store(string(protocol.NodeID(node).Hash()), struct{}{})
// 	visited.Store(string(targetID.Hash()), struct{}{})

// 	var lookups []*lookupBucket

// 	// Start searching for target from α peers closest to T by queuing
// 	// them up and marking them as visited.
// 	for i, peerID := range FindClosestPeers(table, targetID.Hash(), alpha) {
// 		visited.Store(string(peerID.Hash()), struct{}{})

// 		if len(lookups) < numDisjointPaths {
// 			lookups = append(lookups, new(lookupBucket))
// 		}

// 		lookup := lookups[i%numDisjointPaths]
// 		lookup.queue = append(lookup.queue, peerID.(ID))

// 		results = append(results, peerID.(ID))
// 	}

// 	var wait sync.WaitGroup
// 	var mutex sync.Mutex

// 	for _, lookup := range lookups {
// 		go func(lookup *lookupBucket) {
// 			mutex.Lock()
// 			results = append(results, lookup.performLookup(node, table, targetID, alpha, visited)...)
// 			mutex.Unlock()

// 			wait.Done()
// 		}(lookup)

// 		wait.Add(1)
// 	}

// 	// Wait until all D parallel lookups have been completed.
// 	wait.Wait()

// 	// Sort resulting peers by XOR distance.
// 	sort.Slice(results, func(i, j int) bool {
// 		return bytes.Compare(xor(results[i].Hash(), targetID.Hash()), xor(results[j].Hash(), targetID.Hash())) == -1
// 	})

// 	// Cut off list of results to only have the routing table focus on the
// 	// BUCKET_SIZE closest peers to the current node.
// 	if len(results) > BucketSize() {
// 		results = results[:BucketSize()]
// 	}

// 	return
// }
