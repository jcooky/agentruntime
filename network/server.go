package network

import (
	"context"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/di"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"slices"
)

type networkServer struct {
	UnsafeAgentNetworkServer

	service Service
}

func (s *networkServer) CheckLive(ctx context.Context, request *CheckLiveRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.service.CheckLive(ctx, request.Names)
}

func (s *networkServer) GetAgentRuntimeInfo(ctx context.Context, req *GetAgentRuntimeInfoRequest) (*GetAgentRuntimeInfoResponse, error) {

	var (
		runtimeInfo []entity.AgentRuntime
		err         error
	)
	if req.GetAll() {
		runtimeInfo, err = s.service.GetAllAgentRuntimeInfo(ctx)
	} else {
		runtimeInfo, err = s.service.GetAgentRuntimeInfo(ctx, req.Names)
	}
	if err != nil {
		return nil, err
	}

	resp := &GetAgentRuntimeInfoResponse{}
	for _, info := range runtimeInfo {
		resp.AgentRuntimeInfo = append(resp.AgentRuntimeInfo, &AgentRuntimeInfo{
			Addr:       info.Addr,
			Secure:     info.Secure,
			AgentNames: []string{info.Name},
		})
	}

	if req.GetAll() {
		return resp, nil
	}

	// Merge AgentRuntimeInfo by Addr
	for i := 0; i < len(resp.AgentRuntimeInfo); i++ {
		for j := i + 1; j < len(resp.AgentRuntimeInfo); j++ {
			if resp.AgentRuntimeInfo[i].Addr != resp.AgentRuntimeInfo[j].Addr {
				continue
			}

			resp.AgentRuntimeInfo[i].AgentNames = append(resp.AgentRuntimeInfo[i].AgentNames, resp.AgentRuntimeInfo[j].AgentNames...)
			resp.AgentRuntimeInfo = slices.Delete(resp.AgentRuntimeInfo, j, j+1)
			j--
		}
	}

	return resp, nil
}

func (s *networkServer) RegisterAgent(ctx context.Context, req *RegisterAgentRequest) (*emptypb.Empty, error) {
	if err := s.service.RegisterAgent(ctx, req.Addr, req.Secure, req.Names); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *networkServer) DeregisterAgent(ctx context.Context, req *DeregisterAgentRequest) (*emptypb.Empty, error) {
	if err := s.service.DeregisterAgent(ctx, req.Names); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

var (
	_                AgentNetworkServer = (*networkServer)(nil)
	ManagerServerKey                    = di.NewKey()
)

func init() {
	di.Register(ManagerServerKey, func(c context.Context, _ di.Env) (any, error) {
		return &networkServer{
			service: di.MustGet[Service](c, ManagerKey),
		}, nil
	})
}
