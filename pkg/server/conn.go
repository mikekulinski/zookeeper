package server

import (
	"errors"
	"fmt"
	"io"

	pbzk "github.com/mikekulinski/zookeeper/proto"
)

func (s *Server) Message(stream pbzk.Zookeeper_MessageServer) error {
	for {
		// TODO: Add a timeout if we don't at least get a heartbeat message.
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Client closed the stream")
			return nil
		}
		if err != nil {
			return err
		}

		mainResponse := &pbzk.ZookeeperResponse{}
		switch m := req.GetMessage().(type) {
		case *pbzk.ZookeeperRequest_Heartbeat:
			err = s.Heartbeat(m.Heartbeat)
		case *pbzk.ZookeeperRequest_Create:
			var resp *pbzk.CreateResponse
			resp, err = s.Create(m.Create)
			mainResponse.Message = &pbzk.ZookeeperResponse_Create{
				Create: resp,
			}
		case *pbzk.ZookeeperRequest_Delete:
			var resp *pbzk.DeleteResponse
			resp, err = s.Delete(m.Delete)
			mainResponse.Message = &pbzk.ZookeeperResponse_Delete{
				Delete: resp,
			}
		case *pbzk.ZookeeperRequest_Exists:
			var resp *pbzk.ExistsResponse
			resp, err = s.Exists(m.Exists)
			mainResponse.Message = &pbzk.ZookeeperResponse_Exists{
				Exists: resp,
			}
		case *pbzk.ZookeeperRequest_GetData:
			var resp *pbzk.GetDataResponse
			resp, err = s.GetData(m.GetData)
			mainResponse.Message = &pbzk.ZookeeperResponse_GetData{
				GetData: resp,
			}
		case *pbzk.ZookeeperRequest_SetData:
			var resp *pbzk.SetDataResponse
			resp, err = s.SetData(m.SetData)
			mainResponse.Message = &pbzk.ZookeeperResponse_SetData{
				SetData: resp,
			}
		case *pbzk.ZookeeperRequest_GetChildren:
			var resp *pbzk.GetChildrenResponse
			resp, err = s.GetChildren(m.GetChildren)
			mainResponse.Message = &pbzk.ZookeeperResponse_GetChildren{
				GetChildren: resp,
			}
		case *pbzk.ZookeeperRequest_Sync:
			var resp *pbzk.SyncResponse
			resp, err = s.Sync(m.Sync)
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
	}
}

func (s *Server) Heartbeat(_ *pbzk.Heartbeat) error {
	// TODO: Implement some sort of timer reset here.
	return nil
}
