package cluster

/*
See http://stackoverflow.com/questions/900697/how-to-find-the-largest-udp-packet-i-can-send-without-fragmenting

Empirically tested up to 6000 on loopback interface, however real-network
testing was not done

Limiting block size to 512 to guarantee packet delivery
 */
const BlockSize = 512

type Client struct {
    Node    *Node
    Cluster *Cluster
}

func NewClient(name string, partitions int, bind string, port int, existing []string) (*Client, error) {
    node := NewNode(name)
    node.Bind = bind
    node.Port = port

    node.AnnouncePresence()

    cluster, err := NewVia(node, partitions)

    if err == nil && len(existing) > 0 {
        cluster.Join(existing)
    }
    //
    //go func() {
    //    for {
    //        select {
    //        case peer := <- node.Joined:
    //            if *peer.Name == *node.Name {
    //                // stupid but anyways: once I notice I joined, I connect to cluster :D
    //            }
    //            log.Printf("%s joined with %d partitions", *peer.Name, peer.Partitions)
    //        case peer := <- node.Left:
    //            log.Println(*peer.Name, "left")
    //        }
    //    }
    //}()

    return &Client{
        Node: node,
        Cluster: cluster,
    }, err
}

func (client *Client) GetName() string {
    return *client.Node.Name;
}

func (client *Client) Ping(node string) int {
   return client.Cluster.Ping(node)
}

func (client *Client) Load(key []byte, consistencyLevel ConsistencyLevel) ([]byte, int) {
   return client.Cluster.Load(key, consistencyLevel)
}


func (client *Client) Store(key []byte, data []byte, consistencyLevel ConsistencyLevel) int {
   if len(data) > BlockSize {
       // load is limited to block size
       return STORE_ERROR
   }
   return client.Cluster.Store(key, data, consistencyLevel)
}

func (client *Client) KeyNodes(key []byte, consistencyLevel ConsistencyLevel) []*Peer {
    return client.Cluster.HashNodes(key, consistencyLevel)
}

func (client *Client) IsMember() bool {
    return client.Node.IsOperational() && client.Node.server.NumNodes() > 1
}

func (client *Client) DiscoverPeers() []*Peer {
   return client.Cluster.Peers()
}

func (client *Client) Partitions() []*PeerPartition {
   return client.Cluster.Partitions()
}

func (client *Client) Leave() {
    client.Cluster.Disconnect()
}

func (client *Client) Join(existing []string) {
    client.Cluster.Join(existing)
}