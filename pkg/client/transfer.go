package client

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/zzpu/tron-sdk/pkg/common"
	"github.com/zzpu/tron-sdk/pkg/proto/api"
	"github.com/zzpu/tron-sdk/pkg/proto/core"
)

// Transfer from to base58 address
func (g *GrpcClient) Transfer(from, toAddress string, amount int64) (*api.TransactionExtention, error) {
	var err error

	contract := &core.TransferContract{}
	if contract.OwnerAddress, err = common.DecodeCheck(from); err != nil {
		return nil, err
	}
	if contract.ToAddress, err = common.DecodeCheck(toAddress); err != nil {
		return nil, err
	}
	contract.Amount = amount

	ctx, cancel := context.WithTimeout(context.Background(), g.grpcTimeout)
	defer cancel()

	tx, err := g.Client.CreateTransaction2(ctx, contract)
	if err != nil {
		return nil, err
	}
	if proto.Size(tx) == 0 {
		return nil, fmt.Errorf("bad transaction")
	}
	if tx.GetResult().GetCode() != 0 {
		return nil, fmt.Errorf("%s", tx.GetResult().GetMessage())
	}
	return tx, nil
}
