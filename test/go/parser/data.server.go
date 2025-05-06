package parser

import (
	"context"
	"encoding/json"
	"github.com/wskfjtheqian/hbuf_golang/pkg/erro"
	"github.com/wskfjtheqian/hbuf_golang/pkg/hbuf"
	"github.com/wskfjtheqian/hbuf_golang/pkg/manage"
	"github.com/wskfjtheqian/hbuf_golang/pkg/rpc"
)

// UserServer 12
type UserServer interface {
	Init(ctx context.Context)

	//GetInfo 12
	GetInfo(ctx context.Context, req *GetInfoReq) (*GetInfoResp, error)
}

type UserServerClient struct {
	client rpc.Client
}

func (p *UserServerClient) Init(ctx context.Context) {
}

func (p *UserServerClient) GetName() string {
	return "user_server"
}

func (p *UserServerClient) GetId() uint32 {
	return 1
}

func NewUserServerClient(client rpc.Client) *UserServerClient {
	return &UserServerClient{
		client: client,
	}
}

// GetInfo 12
func (r *UserServerClient) GetInfo(ctx context.Context, req *GetInfoReq) (*GetInfoResp, error) {
	ret, err := r.client.Invoke(ctx, req, "user_server/user_server/get_info", &rpc.ClientInvoke{
		ToData: func(buf []byte) (hbuf.Data, error) {
			var req GetInfoResp
			return &req, json.Unmarshal(buf, &req)
		},
		FormData: func(data hbuf.Data) ([]byte, error) {
			return json.Marshal(&data)
		},
	}, 1, &rpc.ClientInvoke{})
	if err != nil {
		return nil, err
	}
	return ret.(*GetInfoResp), nil
}

type UserServerRouter struct {
	server UserServer
	names  map[string]*rpc.ServerInvoke
}

func (p *UserServerRouter) GetName() string {
	return "user_server"
}

func (p *UserServerRouter) GetId() uint32 {
	return 1
}

func (p *UserServerRouter) GetServer() rpc.Init {
	return p.server
}

func (p *UserServerRouter) GetInvoke() map[string]*rpc.ServerInvoke {
	return p.names
}

func NewUserServerRouter(server UserServer) *UserServerRouter {
	return &UserServerRouter{
		server: server,
		names: map[string]*rpc.ServerInvoke{
			"user_server/get_info": {
				ToData: func(buf []byte) (hbuf.Data, error) {
					var req GetInfoReq
					return &req, json.Unmarshal(buf, &req)
				},
				FormData: func(data hbuf.Data) ([]byte, error) {
					return json.Marshal(&data)
				},
				SetInfo: func(ctx context.Context) {
				},
				Invoke: func(ctx context.Context, data hbuf.Data) (hbuf.Data, error) {
					return server.GetInfo(ctx, data.(*GetInfoReq))
				},
			},
		},
	}
}

// UserServer 12
type DefaultUserServer struct {
}

func (s *DefaultUserServer) Init(ctx context.Context) {
}

// GetInfo 12
func (s *DefaultUserServer) GetInfo(ctx context.Context, req *GetInfoReq) (*GetInfoResp, error) {
	return nil, erro.NewError("not find server user_server")
}

var NotFoundUserServer = &DefaultUserServer{}

func GetUserServer(ctx context.Context) UserServer {
	router := manage.GET(ctx).Get(&UserServerRouter{})
	if nil == router {
		return NotFoundUserServer
	}
	if val, ok := router.(UserServer); ok {
		return val
	}
	return NotFoundUserServer
}
