package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
)

type GossipManager struct {
	memberlist *memberlist.Memberlist
	config     *config
	r          RaftNode
}

type myEvents struct {
	gm *GossipManager
}

func (e *myEvents) NotifyJoin(n *memberlist.Node) {
	e.gm.AddRaftNode(n)

	fmt.Printf("%s is activating notifyjoin \n", n.Name)
}
func (e *myEvents) NotifyLeave(n *memberlist.Node) {
	// leave the raft here
	// add later
	fmt.Printf("%s is activating notifyleave \n", n.Name)
}
func (e *myEvents) NotifyUpdate(n *memberlist.Node) {
	fmt.Println("this isn't really supposed to happen???")
	fmt.Printf("%s is activating notifyupdate \n", n.Name)
}

type myDelegate struct {
	raftport string
}

func (d *myDelegate) NodeMeta(_ int) []byte {
	return []byte(d.raftport)
}

func (d *myDelegate) GetBroadcasts(overhead, limit int) [][]byte { return [][]byte{} }
func (d *myDelegate) LocalState(join bool) []byte                { return []byte{} }
func (d *myDelegate) MergeRemoteState(buf []byte, join bool)     {}
func (d *myDelegate) NotifyMsg([]byte)                           {}

// NewGossipManager creates a new gossip manager
func NewGossipManager(cfg *config) (*GossipManager, error) {
	config := memberlist.DefaultLANConfig()
	port, _ := strconv.Atoi(cfg.gossipPort)

	mynewDelegate := &myDelegate{raftport: cfg.raftPort}

	config.Name = cfg.id
	config.BindPort = port
	config.AdvertisePort = port
	config.Delegate = mynewDelegate

	list, err := memberlist.Create(config)

	if err != nil {
		return nil, err
	}

	out := &GossipManager{
		memberlist: list,
		config:     cfg,
		r:          nil,
	}

	config.Events = &myEvents{gm: out}

	return out, nil
}

func (gm *GossipManager) SetRaftNode(raftNode RaftNode) {
	gm.r = raftNode
}

func (gm *GossipManager) AddRaftNode(node *memberlist.Node) {
	if gm.r == nil {
		return
	}

	fmt.Printf("The data we need is name: %s address: %s and port: %d\n", node.Name, node.Addr, node.Port)
	if gm.r.State() != raft.Leader {
		return
	}

	fmt.Printf("node.Meta: %v\n", node.Meta)

	newRaftPort := string(node.Meta)

	raftAddr := fmt.Sprintf("%s:%s", node.Addr.String(), newRaftPort)
	fmt.Printf("Thea ddress we do have is: %s and the node name is: %s\n", raftAddr, node.Name)

	err := gm.r.AddVoter(raft.ServerID(node.Name), raft.ServerAddress(raftAddr), 0, 0).Error()
	if err != nil {
		panic(fmt.Sprintf("failure when connecting to raft: %s", err))
	}
	fmt.Printf("Successfully connected over raft!!!!!!!!!!!!!!!!!!")
}

func (gm *GossipManager) JoinCluster(existing []string) error { // this is where you actually join
	n, err := gm.memberlist.Join(existing)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully joined cluster. Connected to %d nodes\n", n)
	return nil
}

func (gm *GossipManager) GetMembers() []*memberlist.Node {
	return gm.memberlist.Members()
}

func (gm *GossipManager) Leave() error {
	return gm.memberlist.Leave(time.Second * 5)
}

func (gm *GossipManager) Shutdown() error {
	return gm.memberlist.Shutdown()
}
