package server

import (
	"context"
	"fmt"

	"github.com/mikekulinski/zookeeper/pkg/session"
	pbzk "github.com/mikekulinski/zookeeper/proto"
)

func (s *Server) Connect(_ context.Context, req *pbzk.ConnectRequest) (*pbzk.ConnectResponse, error) {
	if _, ok := s.sessions[req.GetClientId()]; ok {
		return nil, fmt.Errorf("session already exists for this clientID")
	}
	sess := &session.Session{}
	s.sessions[req.GetClientId()] = sess
	return &pbzk.ConnectResponse{}, nil
}

func (s *Server) Close(_ context.Context, req *pbzk.CloseRequest) (*pbzk.CloseResponse, error) {
	delete(s.sessions, req.GetClientId())
	return &pbzk.CloseResponse{}, nil
}
