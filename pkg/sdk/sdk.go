package sdk

import (
	"context"

	"github.com/renanqts/openwrt-sdk/pkg/lucirpc"
)

//go:generate mockgen -destination=../../internal/mocks/openwrt/lucirpc.go -package=mocks . LuciRPC
type LuciRPC interface {
	Uci(context.Context, string, []string) (string, error)
}

// OpenWRT represents an OpenWRT SDK client
type OpenWRT struct {
	lucirpc LuciRPC
}

// New creates a new OpenWRT SDK client
func New(addr, username, password string, rpcID int, insecureSkipVerify bool) (*OpenWRT, error) {
	lrcp, err := lucirpc.New(addr, username, password, rpcID, insecureSkipVerify)
	if err != nil {
		return nil, err
	}

	return &OpenWRT{
		lucirpc: lrcp,
	}, nil
}
