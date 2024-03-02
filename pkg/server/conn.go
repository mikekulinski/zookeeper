package server

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/mikekulinski/zookeeper/pkg/session"
	pbzk "github.com/mikekulinski/zookeeper/proto"
)

func (s *Server) SendMessage(ctx context.Context, stream pbzk.Zookeeper_MessageServer) error {
	for {
		// TODO: Add a timeout if we don't at least get a heartbeat message.
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		mainResponse := &pbzk.ZookeeperResponse{}
		switch m := req.GetMessage().(type) {
		case *pbzk.ZookeeperRequest_Connect:
			err = s.Connect(ctx, m.Connect)
		case *pbzk.ZookeeperRequest_Heartbeat:
			err = s.Heartbeat(ctx, m.Heartbeat)
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
			return nil
		}

		if err != nil {
			return err
		}

		err = stream.Send(mainResponse)
		if err != nil {
			return err
		}
	}
}

func (s *Server) Connect(_ context.Context, req *pbzk.ConnectRequest) error {
	if _, ok := s.sessions[req.GetClientId()]; ok {
		return fmt.Errorf("session already exists for this clientID")
	}
	sess := &session.Session{}
	s.sessions[req.GetClientId()] = sess
	return nil
}

func (s *Server) Heartbeat(_ context.Context, _ *pbzk.Heartbeat) error {
	// TODO: Implement some sort of timer reset here.
	return nil
}
