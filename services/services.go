package services

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync/atomic"

	kms "cloud.google.com/go/kms/apiv1"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/evm-layer2/selaginella/bindings"
	"github.com/evm-layer2/selaginella/common/retry"
	"github.com/evm-layer2/selaginella/config"
	"github.com/evm-layer2/selaginella/database"
	node "github.com/evm-layer2/selaginella/eth_client"
	"github.com/evm-layer2/selaginella/protobuf/pb"
	"github.com/evm-layer2/selaginella/sign"
)

const MaxRecvMessageSize = 1024 * 1024 * 300

type RpcServerConfig struct {
	GrpcHostname string
	GrpcPort     string
}

type HsmConfig struct {
	EnableHsm  bool
	HsmAPIName string
	HsmCreden  string
	HsmAddress string
}

type RpcServer struct {
	*RpcServerConfig
	*HsmConfig
	db                  *database.DB
	ethClients          map[uint64]node.EthClient
	RawL1BridgeContract *bind.BoundContract
	RawL2BridgeContract map[uint64]*bind.BoundContract
	L1BridgeContract    *bindings.L1PoolManager
	L2BridgeContract    map[uint64]*bindings.L2PoolManager
	EthAddress          map[uint64]common.Address
	WEthAddress         map[uint64]common.Address
	pb.UnimplementedBridgeServiceServer
	stopped atomic.Bool
	pb.BridgeServiceServer
	privateKey *ecdsa.PrivateKey
	l1ChainID  uint64
}

func (s *RpcServer) Stop(ctx context.Context) error {
	s.stopped.Store(true)
	return nil
}

func (s *RpcServer) Stopped() bool {
	return s.stopped.Load()
}

func NewRpcServer(ctx context.Context, db *database.DB, grpcCfg *RpcServerConfig, hsmCfg *HsmConfig, chainRpcCfg []*config.RPC, priKey *ecdsa.PrivateKey, l1ChainID uint64) (*RpcServer, error) {
	ethClients := make(map[uint64]node.EthClient)
	var rawL1BridgeContract *bind.BoundContract
	rawL2BridgeContracts := make(map[uint64]*bind.BoundContract)
	var l1BridgeContract *bindings.L1PoolManager
	l2BridgeContracts := make(map[uint64]*bindings.L2PoolManager)
	EthAddress := make(map[uint64]common.Address)
	WEthAddress := make(map[uint64]common.Address)
	for i := range chainRpcCfg {
		if chainRpcCfg[i].ChainId == l1ChainID {
			client, err := node.DialEthClient(ctx, chainRpcCfg[i].RpcUrl)
			if err != nil {
				log.Error("dial client fail", "error", err.Error())
				return nil, err
			}
			ethClients[chainRpcCfg[i].ChainId] = client

			l1Parsed, err := abi.JSON(strings.NewReader(
				bindings.L1PoolManagerABI,
			))
			if err != nil {
				log.Error("selaginella parse l1 pool contract abi fail", "err", err)
				return nil, err
			}

			l1Client, _ := node.DialEthClientWithTimeout(ctx, chainRpcCfg[i].RpcUrl, false)
			rawL1PoolContract := bind.NewBoundContract(
				common.HexToAddress(chainRpcCfg[i].FoundingPoolAddress), l1Parsed, l1Client, l1Client,
				l1Client,
			)
			rawL1BridgeContract = rawL1PoolContract

			l1PoolContract, err := bindings.NewL1PoolManager(common.HexToAddress(chainRpcCfg[i].FoundingPoolAddress), l1Client)
			l1BridgeContract = l1PoolContract
			EthAddress[chainRpcCfg[i].ChainId] = common.HexToAddress(chainRpcCfg[i].EthAddress)
			WEthAddress[chainRpcCfg[i].ChainId] = common.HexToAddress(chainRpcCfg[i].WEthAddress)

		} else {
			client, err := node.DialEthClient(ctx, chainRpcCfg[i].RpcUrl)
			if err != nil {
				log.Error("dial client fail", "error", err.Error())
				return nil, err
			}
			ethClients[chainRpcCfg[i].ChainId] = client

			l2Parsed, err := abi.JSON(strings.NewReader(
				bindings.L2PoolManagerABI,
			))
			if err != nil {
				log.Error("selaginella parse l2 pool contract abi fail", "err", err)
				return nil, err
			}

			l2Client, _ := node.DialEthClientWithTimeout(ctx, chainRpcCfg[i].RpcUrl, false)
			rawL2PoolContract := bind.NewBoundContract(
				common.HexToAddress(chainRpcCfg[i].FoundingPoolAddress), l2Parsed, l2Client, l2Client,
				l2Client,
			)
			rawL2BridgeContracts[chainRpcCfg[i].ChainId] = rawL2PoolContract

			l2PoolContract, err := bindings.NewL2PoolManager(common.HexToAddress(chainRpcCfg[i].FoundingPoolAddress), l2Client)
			l2BridgeContracts[chainRpcCfg[i].ChainId] = l2PoolContract
			EthAddress[chainRpcCfg[i].ChainId] = common.HexToAddress(chainRpcCfg[i].EthAddress)
			WEthAddress[chainRpcCfg[i].ChainId] = common.HexToAddress(chainRpcCfg[i].WEthAddress)

		}

	}

	return &RpcServer{
		RpcServerConfig:     grpcCfg,
		db:                  db,
		HsmConfig:           hsmCfg,
		ethClients:          ethClients,
		RawL1BridgeContract: rawL1BridgeContract,
		RawL2BridgeContract: rawL2BridgeContracts,
		L1BridgeContract:    l1BridgeContract,
		L2BridgeContract:    l2BridgeContracts,
		EthAddress:          EthAddress,
		WEthAddress:         WEthAddress,
		privateKey:          priKey,
		l1ChainID:           l1ChainID,
	}, nil
}

func (s *RpcServer) Start(ctx context.Context) error {
	go func(s *RpcServer) {
		addr := fmt.Sprintf("%s:%s", s.GrpcHostname, s.GrpcPort)
		log.Info("start rpc server", "addr", addr)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("Could not start tcp listener. ")
		}

		opt := grpc.MaxRecvMsgSize(MaxRecvMessageSize)

		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(),
		)
		reflection.Register(gs)
		pb.RegisterBridgeServiceServer(gs, s)

		log.Info("grpc info", "port", s.GrpcPort, "address", listener.Addr().String())
		if err := gs.Serve(listener); err != nil {
			log.Error("Could not GRPC server")
		}
	}(s)
	return nil
}

func (s *RpcServer) CrossChainTransfer(ctx context.Context, in *pb.CrossChainTransferRequest) (*pb.CrossChainTransferResponse, error) {
	if in == nil {
		log.Warn("invalid request: request body is empty")
		return nil, errors.New("invalid request: request body is empty")
	}

	var opts *bind.TransactOpts
	var err error
	var tx *types.Transaction
	var finalTx *types.Transaction
	var crossChainTransfers []database.CrossChainTransfer

	for chainId, client := range s.ethClients {
		sourceChainId, _ := new(big.Int).SetString(in.SourceChainId, 10)
		destChainId, _ := new(big.Int).SetString(in.DestChainId, 10)
		amount, _ := new(big.Int).SetString(in.Amount, 10)
		fee, _ := new(big.Int).SetString(in.Fee, 10)
		nonce, _ := new(big.Int).SetString(in.Nonce, 10)
		if chainId == destChainId.Uint64() {
			if s.EnableHsm {
				seqBytes, err := hex.DecodeString(s.HsmCreden)
				if err != nil {
					log.Crit("selaginella", "decode hsm creden fail", err.Error())
				}
				apikey := option.WithCredentialsJSON(seqBytes)
				kClient, err := kms.NewKeyManagementClient(context.Background(), apikey)
				if err != nil {
					log.Crit("selaginella", "create signer error", err.Error())
				}
				mk := &sign.ManagedKey{
					KeyName:      s.HsmAPIName,
					EthereumAddr: common.HexToAddress(s.HsmAddress),
					Gclient:      kClient,
				}
				opts, err = mk.NewEthereumTransactorWithChainID(context.Background(), new(big.Int).SetUint64(chainId))
				if err != nil {
					log.Crit("selaginella", "create signer error", err.Error())
				}
			} else {
				if s.privateKey == nil {
					log.Crit("selaginella", "create signer error", err.Error())
					return nil, errors.New("no private key provided")
				}

				opts, err = bind.NewKeyedTransactorWithChainID(s.privateKey, new(big.Int).SetUint64(chainId))
				if err != nil {
					log.Error("new keyed transactor fail", "err", err)
					return nil, err
				}
			}

			switch in.TokenAddress {
			case s.EthAddress[chainId].String():
				if chainId == s.l1ChainID {
					tx, err = s.L1BridgeContract.BridgeFinalizeETH(opts, sourceChainId, destChainId, common.HexToAddress(in.ReceiveAddress), amount, fee, nonce)
					if err != nil {
						log.Error("get bridge transaction by abi fail", "error", err)
						return nil, err
					}
				} else {
					tx, err = s.L2BridgeContract[chainId].BridgeFinalizeETH(opts, sourceChainId, destChainId, common.HexToAddress(in.ReceiveAddress), amount, fee, nonce)
					if err != nil {
						log.Error("get bridge transaction by abi fail", "error", err)
						return nil, err
					}
				}
			case s.WEthAddress[chainId].String():
				if chainId == s.l1ChainID {
					tx, err = s.L1BridgeContract.BridgeFinalizeWETH(opts, sourceChainId, destChainId, common.HexToAddress(in.ReceiveAddress), amount, fee, nonce)
					if err != nil {
						log.Error("get bridge transaction by abi fail", "error", err)
						return nil, err
					}
				} else {
					tx, err = s.L2BridgeContract[chainId].BridgeFinalizeWETH(opts, sourceChainId, destChainId, common.HexToAddress(in.ReceiveAddress), amount, fee, nonce)
					if err != nil {
						log.Error("get bridge transaction by abi fail", "error", err)
						return nil, err
					}
				}

			default:
				if chainId == s.l1ChainID {
					tx, err = s.L1BridgeContract.BridgeFinalizeERC20(opts, sourceChainId, destChainId, common.HexToAddress(in.ReceiveAddress), common.HexToAddress(in.TokenAddress), amount, fee, nonce)
					if err != nil {
						log.Error("get bridge transaction by abi fail", "error", err)
						return nil, err
					}
				} else {
					tx, err = s.L2BridgeContract[chainId].BridgeFinalizeERC20(opts, sourceChainId, destChainId, common.HexToAddress(in.ReceiveAddress), common.HexToAddress(in.TokenAddress), amount, fee, nonce)
					if err != nil {
						log.Error("get bridge transaction by abi fail", "error", err)
						return nil, err
					}
				}
			}

			finalTx, err = s.RawL2BridgeContract[chainId].RawTransact(opts, tx.Data())
			if err != nil {
				log.Error("raw send bridge transaction fail", "error", err)
				return nil, err
			}

			err = client.SendTransaction(ctx, finalTx)
			if err != nil {
				log.Error("send bridge transaction fail", "error", err)
			}

			receipt, err := client.TxReceiptDetailByHash(finalTx.Hash())
			if err != nil {
				log.Warn("get transaction by hash fail", "err", err)
			}
			log.Info("send bridge transaction success", "tx hash", receipt.TxHash)

			crossChainTransfer := s.db.CrossChainTransfer.BuildCrossChainTransfer(in, finalTx.Hash())
			crossChainTransfers = append(crossChainTransfers, crossChainTransfer)
		}

		retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
		if _, err := retry.Do[interface{}](ctx, 10, retryStrategy, func() (interface{}, error) {
			if err := s.db.Transaction(func(tx *database.DB) error {
				if len(crossChainTransfers) > 0 {
					if err := s.db.CrossChainTransfer.StoreBatchCrossChainTransfer(crossChainTransfers); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				log.Error("unable to persist batch", "err", err)
				return nil, fmt.Errorf("unable to persist batch: %w", err)
			}
			return nil, nil
		}); err != nil {
			return nil, err
		}
	}

	return &pb.CrossChainTransferResponse{
		Success: true,
		Message: "call cross chain transfer success",
	}, nil
}

func (s *RpcServer) ChangeTransferStatus(ctx context.Context, in *pb.CrossChainTransferStatusRequest) (*pb.CrossChainTransferStatusResponse, error) {
	if in == nil {
		log.Warn("invalid request: request body is empty")
		return nil, errors.New("invalid request: request body is empty")
	}

	err := s.db.CrossChainTransfer.ChangeCrossChainTransferStatueByTxHash(in.TxHash)
	if err != nil {
		log.Error("change cross chain transfer status fail", "err", err)
		return nil, err
	}

	return &pb.CrossChainTransferStatusResponse{
		Success: true,
		Message: "call cross chain transfer success",
	}, nil
}
