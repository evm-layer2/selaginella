package etherman

import (
	"math/big"

	"github.com/stretchr/testify/mock"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ EthClient = &MockEthClient{}

type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) TxReceiptByHash(hash common.Hash) (*types.Receipt, error) {
	args := m.Called(hash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockEthClient) BlockHeaderByNumber(number *big.Int) (*types.Header, error) {
	args := m.Called(number)
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEthClient) BlockHeaderByHash(hash common.Hash) (*types.Header, error) {
	args := m.Called(hash)
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEthClient) BlockHeadersByRange(from, to *big.Int) ([]types.Header, error) {
	args := m.Called(from, to)
	return args.Get(0).([]types.Header), args.Error(1)
}

func (m *MockEthClient) TxByHash(hash common.Hash) (*types.Transaction, error) {
	args := m.Called(hash)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockEthClient) StorageHash(address common.Address, blockNumber *big.Int) (common.Hash, error) {
	args := m.Called(address, blockNumber)
	return args.Get(0).(common.Hash), args.Error(1)
}

func (m *MockEthClient) FilterLogs(query ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(query)
	return args.Get(0).([]types.Log), args.Error(1)
}
