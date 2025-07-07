package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

type GossipManager struct {
	memberlist *memberlist.Memberlist
	config     *config
}

// NewGossipManager creates a new gossip manager
func NewGossipManager(cfg *config) (*GossipManager, error) {
	config := memberlist.DefaultLocalConfig()
	port, _ := strconv.Atoi(cfg.gossipPort)

	config.Name = cfg.id
	config.BindPort = port
	config.AdvertisePort = port

	list, err := memberlist.Create(config)

	if err != nil {
		return nil, err
	}

	return &GossipManager{
		memberlist: list,
		config:     cfg,
	}, nil
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
