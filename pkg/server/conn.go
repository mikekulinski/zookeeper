package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mikekulinski/zookeeper/pkg/session"
	"github.com/mikekulinski/zookeeper/pkg/utils"
	pbzk "github.com/mikekulinski/zookeeper/proto"
)

func (s *Server) Message(stream pbzk.Zookeeper_MessageServer) error {
	ctx := stream.Context()
	// Extract the clientID from the message headers.
	clientID, ok := utils.ExtractClientIDHeader(ctx)
	if !ok {
		return fmt.Errorf("missing ClientID in the headers")
	}

	// Establish a new session so that we have a channel we can use to safely process messages.
	sess, err := s.StartSession(clientID)
	if err != nil {
		return fmt.Errorf("error starting session: %w", err)
	}
	defer s.CloseSession(ctx)

	go s.continuouslyReceiveMessages(sess, stream)

	for {
		select {
		case m := <-sess.Messages:
			var resp *pbzk.ZookeeperResponse
			if m.ClientRequest != nil {
				var err error
				resp, err = s.handleClientRequest(ctx, m.ClientRequest)
				if err != nil {
					return err
				}
			} else if m.WatchEvent != nil {
				resp = s.handleWatchEvent(m.WatchEvent)
			} else if m.EOF {
				// There are no more messages so safely close the connection.
				return nil
			}

			// Send the response back to the client.
			err = stream.Send(resp)
			if err != nil {
				return err
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timed out waiting for message to process")
		}
	}
}

// TODO: Route each write request to the request processor that returns a transaction.
func (s *Server) handleClientRequest(ctx context.Context, req *pbzk.ZookeeperRequest) (*pbzk.ZookeeperResponse, error) {
	mainResponse := &pbzk.ZookeeperResponse{}
	var err error
	switch m := req.GetMessage().(type) {
	case *pbzk.ZookeeperRequest_Heartbeat:
		var resp *pbzk.HeartbeatResponse
		resp, err = s.Heartbeat(m.Heartbeat)
		mainResponse.Message = &pbzk.ZookeeperResponse_Heartbeat{
			Heartbeat: resp,
		}
		log.Println("Sending heartbeat response")
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
		return nil, fmt.Errorf("invalid message format: %+v", m)
	}

	if err != nil {
		return nil, fmt.Errorf("error handling client request: %w", err)
	}
	return mainResponse, nil
}

func (s *Server) handleWatchEvent(event *pbzk.WatchEvent) *pbzk.ZookeeperResponse {
	return &pbzk.ZookeeperResponse{
		Message: &pbzk.ZookeeperResponse_WatchEvent{
			WatchEvent: event,
		},
	}
}

func (s *Server) Heartbeat(_ *pbzk.HeartbeatRequest) (*pbzk.HeartbeatResponse, error) {
	// TODO: Implement some sort of timer reset here.
	return &pbzk.HeartbeatResponse{
		ReceivedTsMs: time.Now().UnixMilli(),
	}, nil
}

func (s *Server) StartSession(clientID string) (*session.Session, error) {
	if _, ok := s.sessions[clientID]; ok {
		return nil, fmt.Errorf("session already exists for that clientID")
	}

	sess := session.NewSession()
	s.sessions[clientID] = sess
	return sess, nil
}

func (s *Server) CloseSession(ctx context.Context) {
	// Delete all ephemeral nodes associated with this session.
	clientID, _ := utils.ExtractClientIDHeader(ctx)
	if sess, ok := s.sessions[clientID]; ok {
		for path, node := range sess.EphemeralNodes {
			req := &pbzk.DeleteRequest{
				Path:    path,
				Version: node.Version,
			}
			// Try deleting the node from the tree. This will also clean up the reference to that node in this session.
			// This is ok because it is safe to delete an entry from a map while iterating through it in Go.
			_, err := s.Delete(ctx, req)
			if err != nil {
				panic("unrecoverable: error deleting the ephemeral nodes from tree")
			}
		}
	}
	// Then actually delete the session.
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
			return
		}
		if err != nil {
			log.Printf("Error receiving message from server stream: %+v\n", err)
			return
		}
		sess.Messages <- &session.Event{ClientRequest: req}
	}
}
