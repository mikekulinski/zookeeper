package server

import (
	"fmt"

	"github.com/mikekulinski/zookeeper/pkg/session"
	"github.com/mikekulinski/zookeeper/pkg/zookeeper"
)

func (s *Server) Connect(req *zookeeper.ConnectReq, _ *zookeeper.ConnectResp) error {
	if _, ok := s.sessions[req.ClientID]; ok {
		return fmt.Errorf("session already exists for this clientID")
	}
	sess := &session.Session{}
	s.sessions[req.ClientID] = sess
	return nil
}

func (s *Server) Close(req *zookeeper.CloseReq, _ *zookeeper.CloseResp) error {
	delete(s.sessions, req.ClientID)
	return nil
}
