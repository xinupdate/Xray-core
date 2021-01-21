package smux

import (
	"context"
	"github.com/xtaci/smux"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/transport/internet"
	"sync"
)

type ConnectionSessionPair struct {
	connection internet.Connection
	session    *smux.Session
}

type SmuxManager struct {
	access           sync.RWMutex
	muxConnectionMap map[net.Destination]ConnectionSessionPair
}

func NewSmuxManager() *SmuxManager {
	return &SmuxManager{
		muxConnectionMap: make(map[net.Destination]ConnectionSessionPair),
	}
}

func (sm *SmuxManager) addConnection(ctx context.Context, dest net.Destination, dialer internet.Dialer) error {
	conn, err := dialer.Dial(ctx, dest)
	if err != nil {
		return err
	}

	muxSession, err := smux.Client(conn, nil)
	if err != nil {
		return err
	}

	sm.muxConnectionMap[dest] = ConnectionSessionPair{
		conn,
		muxSession,
	}

	return nil
}

func (sm *SmuxManager) getConnection(ctx context.Context, dest net.Destination, dialer internet.Dialer) (*smux.Stream, error) {
	sm.access.Lock()
	defer sm.access.Unlock()

	// Get session id for logging purpose
	sid := session.ExportIDToError(ctx)

	// Check if the connection already exist
	if _, ok := sm.muxConnectionMap[dest]; !ok {
		// Using existing MUX connection
		newError("Creating new MUX connection for ", dest).AtInfo().WriteToLog(sid)

		if err := sm.addConnection(ctx, dest, dialer); err != nil {
			return nil, err
		}
	}

	// Mux session should exist
	var muxSession *smux.Session
	if pair, ok := sm.muxConnectionMap[dest]; !ok {
		return nil, newError("error retrieving existing mux session")
	} else {
		muxSession = pair.session
	}

	// Open new stream
	stream, err := muxSession.OpenStream()
	if err != nil {
		return nil, err
	}

	newError("new mux stream established ", dest).AtInfo().WriteToLog(sid)

	return stream, nil
}

func (sm *SmuxManager) removeConnection(dest net.Destination) {

}
