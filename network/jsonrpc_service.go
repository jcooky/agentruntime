package network

import (
	"net/http"

	"github.com/gorilla/rpc/v2"
	"github.com/habiliai/agentruntime/errors"
	"github.com/jcooky/go-din"

	"github.com/habiliai/agentruntime/entity"
)

type (
	JsonRpcService struct {
		service Service
	}

	CheckLiveRequest struct {
		Names []string `json:"names"`
	}

	DeregisterAgentRequest struct {
		Names []string `json:"names"`
	}

	RegisterAgentRequest struct {
		Addr   string       `json:"addr"`
		Secure bool         `json:"secure,omitempty"`
		Info   []*AgentInfo `json:"info"`
	}

	GetAgentRuntimeInfoRequest struct {
		Names []string `json:"names,omitempty"`
		All   bool     `json:"all,omitempty"`
	}

	GetAgentRuntimeInfoResponse struct {
		AgentRuntimeInfo []*AgentRuntimeInfo `json:"agent_runtime_info,omitempty"`
	}

	AgentInfo struct {
		Name     string            `json:"name"`
		Role     string            `json:"role"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	AgentRuntimeInfo struct {
		Info   *AgentInfo `json:"info"`
		Addr   string     `json:"addr"`
		Secure bool       `json:"secure,omitempty"`
	}
)

func (s *JsonRpcService) CheckLive(r *http.Request, args *CheckLiveRequest, _ *struct{}) error {
	return s.service.CheckLive(r.Context(), args.Names)
}

func (s *JsonRpcService) GetAgentRuntimeInfo(r *http.Request, args *GetAgentRuntimeInfoRequest, reply *GetAgentRuntimeInfoResponse) error {
	var (
		runtimeInfo []entity.AgentRuntime
		err         error
	)
	if args.All {
		runtimeInfo, err = s.service.GetAllAgentRuntimeInfo(r.Context())
	} else {
		runtimeInfo, err = s.service.GetAgentRuntimeInfo(r.Context(), args.Names)
	}
	if err != nil {
		return err
	}

	reply.AgentRuntimeInfo = make([]*AgentRuntimeInfo, 0, len(runtimeInfo))
	for _, agent := range runtimeInfo {
		reply.AgentRuntimeInfo = append(reply.AgentRuntimeInfo, &AgentRuntimeInfo{
			Addr:   agent.Addr,
			Secure: agent.Secure,
			Info: &AgentInfo{
				Name:     agent.Name,
				Role:     agent.Role,
				Metadata: agent.Metadata.Data(),
			},
		})
	}

	return nil
}

func (s *JsonRpcService) RegisterAgent(r *http.Request, args *RegisterAgentRequest, _ *struct{}) error {
	if err := s.service.RegisterAgent(r.Context(), args.Addr, args.Secure, args.Info); err != nil {
		return err
	}

	return nil
}

func (s *JsonRpcService) DeregisterAgent(r *http.Request, args *DeregisterAgentRequest, _ *struct{}) error {
	if err := s.service.DeregisterAgent(r.Context(), args.Names); err != nil {
		return err
	}

	return nil
}

var (
	servicePrefix = "habiliai.agentnetwork.v1"
)

func RegisterJsonRpcService(c *din.Container, server *rpc.Server) error {
	svc := &JsonRpcService{
		service: din.MustGetT[Service](c),
	}

	return errors.Wrapf(server.RegisterService(svc, servicePrefix), "failed to register jsonrpc service")
}
