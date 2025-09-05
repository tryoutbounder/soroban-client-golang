package blend

import (
	"net/http"

	"github.com/tryOutbounder/soroban-client-golang/blend/types/backstop"
	soroban "github.com/tryOutbounder/soroban-client-golang/pkg/rpc"
)

type BlendClient struct {
	rpc *soroban.RpcClient
}

func NewBlendClient(rpcUrl string) *BlendClient {
	return &BlendClient{
		rpc: soroban.NewClient(rpcUrl, http.DefaultClient),
	}
}

// Backstop Data Calls

// Load the configuration of the backstop
func (bc *BlendClient) BackstopConfig(
	backstopAddr string,
) (*backstop.BackstopConfig, error) {
	return backstop.LoadConfig(bc.rpc, backstopAddr)
}

// Load token price, makeup, and analytics
func (bc *BlendClient) BackstopToken(
	cometContract string,
	blndTokenContract string,
	usdcTokenContract string,
) (*backstop.BackstopToken, error) {
	return backstop.LoadToken(
		bc.rpc,
		cometContract,
		blndTokenContract,
		usdcTokenContract,
	)
}

// Load the balance of assets in the pool's backstop
func (bc *BlendClient) BackstopPoolBalance(
	backstopContract string,
	poolContract string,
) (*backstop.BackstopPoolBalance, error) {
	return backstop.LoadPoolBalance(
		bc.rpc,
		backstopContract,
		poolContract,
	)
}

// User balance and info in the pool
func (bc *BlendClient) BackstopPoolUser(
	backstopContract string,
	poolContract string,
	userAddress string,
) (*backstop.BackstopPoolUser, error) {
	return backstop.LoadBackstopUser(
		bc.rpc,
		backstopContract,
		poolContract,
		userAddress,
	)
}

func (bc *BlendClient) BackstopDepositors(backstopContract string, poolContract string) {

}

// Pool Data Calls
// func (bc *BlendClient) PoolMetadata()

// func (bc *BlendClient) PoolEvents()

// func (bc *BlendClient) PoolReserveAddresses()

// func (bc *BlendClient) PoolReserve()

// func (bc *BlendClient) PoolOracle()

// func (bc *BlendClient) PoolUser()
