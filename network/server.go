package network

import (
	"context"
	"github.com/jcooky/go-din"

	"github.com/habiliai/agentruntime/entity"
	"google.golang.org/protobuf/types/known/emptypb"
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
	resp := &GetAgentRuntimeInfoResponse{
		AgentRuntimeInfo: make([]*AgentRuntimeInfo, 0, len(runtimeInfo)),
	}
	for _, agent := range runtimeInfo {
		resp.AgentRuntimeInfo = append(resp.AgentRuntimeInfo, &AgentRuntimeInfo{
			Addr:   agent.Addr,
			Secure: agent.Secure,
			Info: &AgentInfo{
				Name:     agent.Name,
				Role:     agent.Role,
				Metadata: agent.Metadata.Data(),
			},
		})
	}

	return resp, nil
}

func (s *networkServer) RegisterAgent(ctx context.Context, req *RegisterAgentRequest) (*emptypb.Empty, error) {
	if err := s.service.RegisterAgent(ctx, req.Addr, req.Secure, req.Info); err != nil {
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

func init() {
	din.RegisterT(func(c *din.Container) (AgentNetworkServer, error) {
		return &networkServer{
			service: din.MustGetT[Service](c),
		}, nil
	})
}
