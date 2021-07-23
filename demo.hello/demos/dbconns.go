package demos

import (
	"context"
	"fmt"
)

var count uint32

type driverConn string

type connReq struct {
	Conn driverConn
	Err  error
}

type database struct {
	FreeConns []driverConn
	ConnReqs  map[uint32]chan connReq
}

func newDataBase() *database {
	return &database{
		FreeConns: make([]driverConn, 0, 10),
		ConnReqs:  make(map[uint32]chan connReq, 10),
	}
}

func (db *database) PrintInfo() {
	fmt.Printf("Free Connections [count=%d]: %v\n", len(db.FreeConns), db.FreeConns)
	fmt.Printf("Connection Requests: count=%d\n", len(db.ConnReqs))
}

func (db *database) Connect(ctx context.Context) (driverConn, error) {
	if len(db.FreeConns) > 0 {
		conn := db.FreeConns[0]
		db.FreeConns = db.FreeConns[1:]
		return conn, nil
	}

	req := make(chan connReq, 1)
	key := db.getNextReqKey()
	db.ConnReqs[key] = req

	select {
	case <-ctx.Done():
	case ret, ok := <-req:
		if !ok {
			return "", fmt.Errorf("ErrDBClosed")
		}
		return ret.Conn, nil
	}
	panic("not happen")
}

func (db *database) PutConn(conn string) {
	localConn := driverConn(conn)
	if len(db.ConnReqs) == 0 {
		db.FreeConns = append(db.FreeConns, localConn)
		return
	}
	var reqKey uint32
	var req chan connReq
	for reqKey, req = range db.ConnReqs {
		break
	}
	delete(db.ConnReqs, reqKey)
	req <- connReq{
		Conn: localConn,
	}
}

func (db *database) getNextReqKey() uint32 {
	count++
	return count
}
