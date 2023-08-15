package loadbalance_test

import (
	"fmt"
	"net"
	"net/url"
	"testing"

	api "github.com/jhkim988/proglog/api/v1"

	"github.com/jhkim988/proglog/internal/config"
	"github.com/jhkim988/proglog/internal/loadbalance"
	"github.com/jhkim988/proglog/internal/server"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

func TestResolver(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)
	serverCreds := credentials.NewTLS(tlsConfig)

	srv, err := server.NewGRPCServer(&server.Config{
		GetServerer: &getServers{},
	}, grpc.Creds(serverCreds))
	require.NoError(t, err)
	go srv.Serve(l)

	conn := &clientConn{} // Mock
	tlsConfig, err = config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.RootClientCertFile,
		KeyFile:       config.RootClientKeyFile,
		CAFile:        config.CAFile,
		Server:        false,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)
	clientCreds := credentials.NewTLS(tlsConfig)
	opts := resolver.BuildOptions{
		DialCreds: clientCreds,
	}
	r := &loadbalance.Resolver{}
	urlAddr, err := url.Parse(fmt.Sprintf("http://%s", l.Addr().String()))
	require.NoError(t, err)
	_, err = r.Build(
		resolver.Target{
			URL: *urlAddr,
		},
		conn,
		opts,
	)
	require.NoError(t, err)

	wantState := resolver.State{
		Addresses: []resolver.Address{{
			Addr:       "localhost:9001",
			Attributes: attributes.New("is_leader", true),
		}, {
			Addr:       "localhost:9002",
			Attributes: attributes.New("is_leader", false),
		}},
	}
	require.Equal(t, wantState, conn.state)

	conn.state.Addresses = nil
	r.ResolveNow(resolver.ResolveNowOptions{})
	require.Equal(t, wantState, conn.state)
}

// Mock
type getServers struct{}

func (s *getServers) GetServers() ([]*api.Server, error) {
	return []*api.Server{{
		Id:       "leader",
		RpcAddr:  "localhost:9001",
		IsLeader: true,
	}, {
		Id:      "follower",
		RpcAddr: "localhost:9002",
	}}, nil
}

// Mock
type clientConn struct {
	resolver.ClientConn
	state resolver.State
}

func (*clientConn) NewAddress(addresses []resolver.Address) {}

func (*clientConn) NewServiceConfig(serviceConfig string) {}

func (*clientConn) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return nil
}

func (*clientConn) ReportError(error) {}

func (c *clientConn) UpdateState(state resolver.State) error {
	c.state = state
	return nil
}
