// Cluster built on top of node group
package cluster

import (
    "container/list"
    "errors"
    "fmt"
	"github.com/hashicorp/serf/serf"
	"net"
)

type Cluster struct {
    proxy *Node
    storage Storage
    handlers *list.List
    shutdownCh chan int
}

const DefaultPartitions = 127
const PartitionsTag = "partitions"

// Create a cluster instance with node as a communication proxy
func NewVia(node *Node, partitions int) (c *Cluster, err error) {
    if ! node.IsOperational() {
        return nil, errors.New("Node is not ready")
    }

    c = &Cluster{
        proxy: node,
        storage: NewInMemoryStorage(),
        handlers: list.New(),
    }

    c.handlers.PushBack(NewPongActivity(c))
    c.handlers.PushBack(NewBucketStoreActivity(c))
    c.handlers.PushBack(NewBucketLoadActivity(c))
    return c, nil
}

// Returns one primary node and as much consistently determined replication nodes as needed for meeting consistency level
func (c *Cluster) HashNodes(objectHash []byte, level ConsistencyLevel) []*Peer {
    copies := c.Copies(level)
    nodes := make([]*Peer, 0, copies)
    partitions := c.Partitions()

    // first of all determine the primary node
    h, l, r, lenPeers := 0, 0, len(partitions) - 1, len(partitions)
    if Clockwise(partitions[r].Hash(), objectHash, partitions[l].Hash()) {
        h = r
    } else {
        var m int
        for r - l > 1 {
            m = l + (r - l) >> 1

            if Clockwise(partitions[m].Hash(), objectHash, partitions[r].Hash()) {
                l = m
            } else {
                r = m
            }
        }

        h = l
    }

    // build up the array - we take the pivot partition and find as many *other* peers as we need

    for j := 0; len(nodes) < copies; j++ {
        partition := partitions[(h - j + lenPeers) % lenPeers]
        // if partition peer is not in nodes yet - we can add this partition
        peerPresent := false
        for k := 0; k < len(nodes); k++ {
            if partition.Peer == nodes[k] {
                // partition is from already added peer
                peerPresent = true
                break
            }
        }

        if ! peerPresent {
            nodes = append(nodes, partition.Peer)
        }
    }

    return nodes
}

// Ping another cluster peer
func (c *Cluster) Ping(peer string) int {
   activity := NewPingActivity(c)

   e := c.handlers.PushBack(activity)
   defer c.handlers.Remove(e)

   activity.Run(peer)
   return <- activity.Result
}

func (c *Cluster) Store(key, data []byte, level ConsistencyLevel) int {
   activity := NewStoreActivity(c, c.AdjustedConsistencyLevel(level))

   e := c.handlers.PushBack(activity)
   defer c.handlers.Remove(e)

   activity.Run(key, data)
   return <- activity.Result
}

func (c *Cluster) Load(key []byte, level ConsistencyLevel) ([]byte, int) {
   adjustedLevel := c.AdjustedConsistencyLevel(level)
   activity := NewLoadActivity(c, adjustedLevel)

   e := c.handlers.PushBack(activity)
   defer c.handlers.Remove(e)

   activity.Run(key)

   data, result := activity.Data, <- activity.Result

   if result == LOAD_PARTIAL_SUCCESS {
       c.Store(key, data, adjustedLevel)
   }

   return data, result
}

// Send cluster event
func (c *Cluster) Send(name string, m *Message) error {
    return c.proxy.server.UserEvent(name, Marshall(m), true)
}

// Send cluster query
func (c *Cluster) SendQuery(name string, m *Message, param *serf.QueryParam) (*serf.QueryResponse, error) {
	return c.proxy.server.Query(name, Marshall(m), param)
}

func (c *Cluster) Partitions() []*PeerPartition {
    c.proxy.DiscoverPeers()
    peersMap := c.proxy.Peers
    partitions := make([]*PeerPartition, 0, DefaultPartitions*len(peersMap))

    for _, p := range peersMap {
        pp := p.Clone()
        for i := 0; i < p.Partitions; i++ {
            partitions = append(partitions, &PeerPartition{
                Peer: pp,
                Partition: i,
            })
        }
    }

    return PeerPartitionSorterSorter(partitions).ByHash().Sort()
}

func (c *Cluster) Peers() []*Peer {
    // TODO: potential shared memory access
    c.proxy.DiscoverPeers()

    peersMap := c.proxy.Peers
    peers := make([]*Peer, 0, len(peersMap))

    for _, p := range peersMap {
        peers = append(peers, p.Clone())
    }

    return peers
}

func (c *Cluster) Quorum() int {
    return len(c.proxy.Peers) + 1
}

func (c *Cluster) Size() int {
    return len(c.proxy.Peers)
}

// Resolve cluster peer IP address by peer name
func (c *Cluster) GetPeerAddr(peer string) (net.IP, error) {
    p, ok := c.proxy.Peers[peer]
    if !ok {
        return nil, fmt.Errorf("Peer not available: %s", peer)
    }

    return p.Addr, nil
}