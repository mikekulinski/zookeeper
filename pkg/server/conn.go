package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mikekulinski/zookeeper/pkg/session"
	pbzk "github.com/mikekulinski/zookeeper/proto"
)

func (s *Server) Message(stream pbzk.Zookeeper_MessageServer) error {
	// Extract the clientID from the message headers.
	clientID, ok := ExtractClientIDHeader(stream.Context())
	if !ok {
		return fmt.Errorf("missing ClientID in the headers")
	}

	// Establish a new session so that we have a channel we can use to safely process messages.
	sess, err := s.StartSession(clientID)
	if err != nil {
		return fmt.Errorf("error starting session: %w", err)
	}
	defer s.CloseSession(clientID)

	go s.continuouslyReceiveMessages(sess, stream)

	for {
		select {
		case m := <-sess.Messages:
			if m.ClientRequest != nil {
				err := s.handleClientRequest(m.ClientRequest, stream)
				if err != nil {
					return err
				}
			} else if m.WatchEvent != nil {
				err := s.handleWatchEvent(m.WatchEvent, stream)
				if err != nil {
					return err
				}
			} else if m.EOF {
				// There are no more messages so safely close the connection.
				log.Println("Received EOF from messages channel")
				return nil
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timed out waiting for message to process")
		}
	}
}

func (s *Server) handleClientRequest(req *pbzk.ZookeeperRequest, stream pbzk.Zookeeper_MessageServer) error {
	ctx := stream.Context()

	mainResponse := &pbzk.ZookeeperResponse{}
	var err error
	switch m := req.GetMessage().(type) {
	case *pbzk.ZookeeperRequest_Heartbeat:
		err = s.Heartbeat(m.Heartbeat)
	case *pbzk.ZookeeperRequest_Create:
		var resp *pbzk.CreateResponse
		resp, err = s.Create(ctx, m.Create)
		mainResponse.Message = &pbzk.ZookeeperResponse_Create{
			Create: resp,
		}
	case *pbzk.ZookeeperRequest_Delete:
		var resp *pbzk.DeleteResponse
		resp, err = s.Delete(ctx, m.Delete)
		mainResponse.Message = &pbzk.ZookeeperResponse_Delete{
			Delete: resp,
		}
	case *pbzk.ZookeeperRequest_Exists:
		var resp *pbzk.ExistsResponse
		resp, err = s.Exists(ctx, m.Exists)
		mainResponse.Message = &pbzk.ZookeeperResponse_Exists{
			Exists: resp,
		}
	case *pbzk.ZookeeperRequest_GetData:
		var resp *pbzk.GetDataResponse
		resp, err = s.GetData(ctx, m.GetData)
		mainResponse.Message = &pbzk.ZookeeperResponse_GetData{
			GetData: resp,
		}
	case *pbzk.ZookeeperRequest_SetData:
		var resp *pbzk.SetDataResponse
		resp, err = s.SetData(ctx, m.SetData)
		mainResponse.Message = &pbzk.ZookeeperResponse_SetData{
			SetData: resp,
		}
	case *pbzk.ZookeeperRequest_GetChildren:
		var resp *pbzk.GetChildrenResponse
		resp, err = s.GetChildren(ctx, m.GetChildren)
		mainResponse.Message = &pbzk.ZookeeperResponse_GetChildren{
			GetChildren: resp,
		}
	case *pbzk.ZookeeperRequest_Sync:
		var resp *pbzk.SyncResponse
		resp, err = s.Sync(ctx, m.Sync)
		mainResponse.Message = &pbzk.ZookeeperResponse_Sync{
			Sync: resp,
		}
	default:
		return fmt.Errorf("invalid message format: %+v", m)
	}

	if err != nil {
		return err
	}

	err = stream.Send(mainResponse)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleWatchEvent(event *pbzk.WatchEvent, stream pbzk.Zookeeper_MessageServer) error {
	// TODO
	return nil
}

func (s *Server) Heartbeat(_ *pbzk.Heartbeat) error {
	// TODO: Implement some sort of timer reset here.
	return nil
}

func (s *Server) StartSession(clientID string) (*session.Session, error) {
	if _, ok := s.sessions[clientID]; ok {
		return nil, fmt.Errorf("session already exists for that clientID")
	}

	sess := session.NewSession()
	s.sessions[clientID] = sess
	return sess, nil
}

func (s *Server) CloseSession(clientID string) {
	delete(s.sessions, clientID)
}

func (s *Server) continuouslyReceiveMessages(sess *session.Session, stream pbzk.Zookeeper_MessageServer) {
	// Regardless of how we exit this function, we should send an EOF so that we
	// close the stream once we have finished processing all messages.
	defer func() {
		sess.Messages <- &session.Event{EOF: true}
	}()

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("Client closed the stream")
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		sess.Messages <- &session.Event{ClientRequest: req}
	}
}
