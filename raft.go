package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	"github.com/AlexG28/keyvalue/store"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type kvFsm struct {
	store store.Store
}

type kvSnapshot struct {
	store store.Store
}

func (kf *kvFsm) Snapshot() (raft.FSMSnapshot, error) {
	return &kvSnapshot{store: kf.store}, nil
}

func (ks *kvSnapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write([]byte("snapshot"))
	if err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (ks *kvSnapshot) Release() {

}

func (kf *kvFsm) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		var sp setPayload
		if err := json.Unmarshal(log.Data, &sp); err == nil && sp.Key != "" {
			if err := kf.store.Add(sp.Key, sp.Value); err != nil {
				return fmt.Errorf("Could not add to store: %s", err)
			}
			return nil
		}

		var dp deletePayload
		if err := json.Unmarshal(log.Data, &dp); err == nil && dp.Key != "" {
			if err := kf.store.Delete(dp.Key); err != nil {
				return fmt.Errorf("Could not delete from store: %s", err)
			}
			return nil
		}

		return fmt.Errorf("Could not parse payload: unknown operation type")
	default:
		return fmt.Errorf("Unknown raft log type: %#v", log.Type)
	}
}

func (kf *kvFsm) Restore(rc io.ReadCloser) error {
	decoder := json.NewDecoder(rc)

	for decoder.More() {
		var sp setPayload
		err := decoder.Decode(&sp)
		if err != nil {
			return fmt.Errorf("could not decode payload: %s", err)
		}

		if err := kf.store.Add(sp.Key, sp.Value); err != nil {
			return fmt.Errorf("could not restore key-value: %s", err)
		}
	}

	return rc.Close()
}

func setupRaft(dir, nodeId, raftAddress string, kf *kvFsm) (*raft.Raft, error) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create data directory: %s", err)
	}

	store, err := raftboltdb.NewBoltStore(path.Join(dir, "bolt"))
	if err != nil {
		return nil, fmt.Errorf("could not create bolt store: %s", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(path.Join(dir, "snapshot"), 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("could not create snapshot store: %s", err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddress)
	if err != nil {
		return nil, fmt.Errorf("could not resolve address: %s", err)
	}

	transport, err := raft.NewTCPTransport(raftAddress, tcpAddr, 10, time.Second*10, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("could not create tcp transport: %s", err)
	}

	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(nodeId)

	r, err := raft.NewRaft(raftCfg, kf, store, store, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("Could not create raft instance: %s", err)
	}

	r.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeId),
				Address: transport.LocalAddr(),
			},
		},
	})

	return r, nil
}
