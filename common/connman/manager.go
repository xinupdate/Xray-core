package connman

import (
	"context"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/transport/internet"
	"sync"
)

type SmuxManager struct {
	access           sync.RWMutex
	muxConnectionMap map[net.Destination]internet.Connection
}

func NewSmuxManager() *SmuxManager {
	return &SmuxManager{
		muxConnectionMap: make(map[net.Destination]internet.Connection),
	}
}

func (sm *SmuxManager) addConnection(ctx context.Context, dest net.Destination, dialer internet.Dialer) error {
	conn, err := dialer.Dial(ctx, dest)
	if err != nil {
		return err
	}

	sm.muxConnectionMap[dest] = conn
	return nil
}

func (sm *SmuxManager) GetConnection(ctx context.Context, dest net.Destination, dialer internet.Dialer) (internet.Connection, error) {
	sm.access.Lock()
	defer sm.access.Unlock()

	// Check if the connection already exist
	if _, ok := sm.muxConnectionMap[dest]; !ok {
		// Using existing MUX connection
		if err := sm.addConnection(ctx, dest, dialer); err != nil {
			return nil, err
		}
	}

	// Connection should exist
	if conn, ok := sm.muxConnectionMap[dest]; !ok {
		return nil, newError("error retrieving existing mux session")
	} else {
		return conn, nil
	}
}

func (sm *SmuxManager) removeConnection(dest net.Destination) {

}
