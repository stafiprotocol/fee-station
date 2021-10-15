package substrate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"fee-station/shared/substrate/types/stafi"
	wbskt "fee-station/shared/substrate/websocket"

	"github.com/gorilla/websocket"
	scale "github.com/itering/scale.go"
	"github.com/itering/scale.go/source"
	scaleTypes "github.com/itering/scale.go/types"
	"github.com/itering/scale.go/utiles"
	"github.com/itering/substrate-api-rpc/pkg/recws"
	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/itering/substrate-api-rpc/util"
	"github.com/sirupsen/logrus"
	gsrpc "github.com/stafiprotocol/go-substrate-rpc-client"
)

const (
	wsId       = 1
	storageKey = "0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"
)

type SarpcClient struct {
	endpoint           string
	wsPool             wbskt.Pool
	chainType          string
	metaRaw            string
	typesPath          string
	currentSpecVersion int
	metaDecoder        interface{}
}

var (
	metaDecoders = map[string]interface{}{
		ChainTypeStafi:    &stafi.MetadataDecoder{},
		ChainTypePolkadot: &scale.MetadataDecoder{},
	}
)

func NewSarpcClient(chainType, endpoint, typesPath string) (*SarpcClient, error) {
	api, err := gsrpc.NewSubstrateAPI(endpoint)
	if err != nil {
		return nil, err
	}

	latestHash, err := api.RPC.Chain.GetFinalizedHead()
	if err != nil {
		return nil, err
	}
	logrus.Info("NewSarpcClient", "latestHash", latestHash.Hex())

	md := metaDecoders[chainType]
	if md == nil {
		return nil, errors.New("chainType not supported")
	}

	sc := &SarpcClient{
		chainType:          chainType,
		endpoint:           endpoint,
		wsPool:             nil,
		metaRaw:            "",
		typesPath:          typesPath,
		currentSpecVersion: 0,
		metaDecoder:        md,
	}

	sc.regCustomTypes()

	err = sc.UpdateMeta(latestHash.Hex())
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func (sc *SarpcClient) regCustomTypes() {
	content, err := ioutil.ReadFile(sc.typesPath)
	if err != nil {
		panic(err)
	}

	switch sc.chainType {
	case ChainTypeStafi:
		stafi.RuntimeType{}.Reg()
		stafi.RegCustomTypes(source.LoadTypeRegistry(content))
	case ChainTypePolkadot:
		scaleTypes.RuntimeType{}.Reg()
		scaleTypes.RegCustomTypes(source.LoadTypeRegistry(content))
	default:
		panic("chainType not supported")
	}
}

func (sc *SarpcClient) initial() (*wbskt.PoolConn, error) {
	var err error
	if sc.wsPool == nil {
		factory := func() (*recws.RecConn, error) {
			SubscribeConn := &recws.RecConn{KeepAliveTimeout: 10 * time.Second}
			SubscribeConn.Dial(sc.endpoint, nil)
			return SubscribeConn, err
		}
		if sc.wsPool, err = wbskt.NewChannelPool(1, 25, factory); err != nil {
			fmt.Println("NewChannelPool", err)
		}
	}
	if err != nil {
		return nil, err
	}
	conn, err := sc.wsPool.Get()
	return conn, err
}

func (sc *SarpcClient) sendWsRequest(p wbskt.WsConn, v interface{}, action []byte) (err error) {
	if p == nil {
		var pool *wbskt.PoolConn
		if pool, err = sc.initial(); err == nil {
			defer pool.Close()
			p = pool.Conn
		} else {
			return
		}

	}

	if err = p.WriteMessage(websocket.TextMessage, action); err != nil {
		if p != nil {
			p.MarkUnusable()
		}
		return fmt.Errorf("websocket send error: %v", err)
	}
	if err = p.ReadJSON(v); err != nil {
		if p != nil {
			p.MarkUnusable()
		}
		return
	}
	return nil
}

func (sc *SarpcClient) UpdateMeta(blockHash string) error {
	v := &rpc.JsonRpcResult{}
	// runtime version
	if err := sc.sendWsRequest(nil, v, rpc.ChainGetRuntimeVersion(wsId, blockHash)); err != nil {
		return err
	}
	// fmt.Println(fmt.Errorf("%+v", v.Error))
	// fmt.Println(fmt.Errorf("%+v", v.Result))
	if v.Error != nil {
		return fmt.Errorf("%+v", v.Error)
	}
	r := v.ToRuntimeVersion()
	if r == nil {
		return fmt.Errorf("runtime version nil")
	}

	// metadata raw
	if sc.metaRaw == "" || r.SpecVersion > sc.currentSpecVersion {
		if err := sc.sendWsRequest(nil, v, rpc.StateGetMetadata(wsId, blockHash)); err != nil {
			return err
		}
		metaRaw, err := v.ToString()
		if err != nil {
			return err
		}
		sc.metaRaw = metaRaw
		sc.currentSpecVersion = r.SpecVersion

		switch sc.chainType {
		case ChainTypeStafi:
			md, _ := sc.metaDecoder.(*stafi.MetadataDecoder)
			md.Init(utiles.HexToBytes(metaRaw))
			if err := md.Process(); err != nil {
				return err
			}
		case ChainTypePolkadot:
			md, _ := sc.metaDecoder.(*scale.MetadataDecoder)
			md.Init(utiles.HexToBytes(metaRaw))
			if err := md.Process(); err != nil {
				return err
			}
		default:
			return errors.New("chainType not supported")
		}
	}

	return nil
}

func (sc *SarpcClient) GetBlock(blockHash string) (*rpc.Block, error) {
	v := &rpc.JsonRpcResult{}
	if err := sc.sendWsRequest(nil, v, rpc.ChainGetBlock(wsId, blockHash)); err != nil {
		return nil, err
	}
	rpcBlock := v.ToBlock()
	return &rpcBlock.Block, nil
}

func (sc *SarpcClient) GetExtrinsics(blockHash string) ([]*Transaction, error) {
	err := sc.UpdateMeta(blockHash)
	if err != nil {
		return nil, err
	}

	blk, err := sc.GetBlock(blockHash)
	if err != nil {
		return nil, err
	}

	exts := make([]*Transaction, 0)
	switch sc.chainType {
	case ChainTypeStafi:
		e := new(stafi.ExtrinsicDecoder)
		md, _ := sc.metaDecoder.(*stafi.MetadataDecoder)
		option := stafi.ScaleDecoderOption{Metadata: &md.Metadata, Spec: sc.currentSpecVersion}
		for _, raw := range blk.Extrinsics {
			e.Init(stafi.ScaleBytes{Data: util.HexToBytes(raw)}, &option)
			e.Process()
			if e.ExtrinsicHash != "" && e.ContainsTransaction {
				ext := &Transaction{
					ExtrinsicHash:  e.ExtrinsicHash,
					CallModuleName: e.CallModule.Name,
					CallName:       e.Call.Name,
					Address:        e.Address,
					Params:         e.Params,
				}
				exts = append(exts, ext)
			}
		}
		return exts, nil
	case ChainTypePolkadot:
		e := new(scale.ExtrinsicDecoder)
		md, _ := sc.metaDecoder.(*scale.MetadataDecoder)

		option := scaleTypes.ScaleDecoderOption{Metadata: &md.Metadata, Spec: sc.currentSpecVersion}
		for _, raw := range blk.Extrinsics {
			e.Init(scaleTypes.ScaleBytes{Data: util.HexToBytes(raw)}, &option)
			e.Process()

			call, exist := e.Metadata.CallIndex[e.CallIndex]
			if !exist {
				return nil, fmt.Errorf("callIndex: %s not exist metaData", e.CallIndex)
			}

			if e.ExtrinsicHash != "" && e.ContainsTransaction {
				ext := &Transaction{
					ExtrinsicHash:  e.ExtrinsicHash,
					CallModuleName: call.Module.Name,
					CallName:       call.Call.Name,
					Address:        e.Address,
					Params:         e.Params,
				}
				exts = append(exts, ext)
			}
		}
		return exts, nil
	default:
		return nil, errors.New("chainType not supported")
	}
}

func (sc *SarpcClient) GetBlockHash(blockNum uint64) (string, error) {
	v := &rpc.JsonRpcResult{}
	if err := sc.sendWsRequest(nil, v, rpc.ChainGetBlockHash(wsId, int(blockNum))); err != nil {
		return "", fmt.Errorf("websocket get block hash error: %v", err)
	}

	blockHash, err := v.ToString()
	if err != nil {
		return "", err
	}

	return blockHash, nil
}

func (sc *SarpcClient) GetChainEvents(blockHash string) ([]*ChainEvent, error) {
	err := sc.UpdateMeta(blockHash)
	if err != nil {
		return nil, err
	}

	v := &rpc.JsonRpcResult{}
	if err := sc.sendWsRequest(nil, v, rpc.StateGetStorage(wsId, storageKey, blockHash)); err != nil {
		return nil, fmt.Errorf("websocket get event raw error: %v", err)
	}
	eventRaw, err := v.ToString()
	if err != nil {
		return nil, err
	}

	var events []*ChainEvent
	switch sc.chainType {
	case ChainTypeStafi:
		e := stafi.EventsDecoder{}
		md, _ := sc.metaDecoder.(*stafi.MetadataDecoder)
		option := stafi.ScaleDecoderOption{Metadata: &md.Metadata}
		e.Init(stafi.ScaleBytes{Data: util.HexToBytes(eventRaw)}, &option)
		e.Process()
		b, err := json.Marshal(e.Value)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &events)
		if err != nil {
			return nil, err
		}
	case ChainTypePolkadot:
		md, _ := sc.metaDecoder.(*scale.MetadataDecoder)

		option := scaleTypes.ScaleDecoderOption{Metadata: &md.Metadata}
		e := scale.EventsDecoder{}
		e.Init(scaleTypes.ScaleBytes{Data: util.HexToBytes(eventRaw)}, &option)
		e.Process()
		b, err := json.Marshal(e.Value)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &events)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("chainType not supported")
	}

	return events, nil
}

func (sc *SarpcClient) GetEvents(blockNum uint64) ([]*ChainEvent, error) {
	blockHash, err := sc.GetBlockHash(blockNum)
	if err != nil {
		return nil, err
	}

	evts, err := sc.GetChainEvents(blockHash)
	if err != nil {
		return nil, err
	}

	return evts, nil
}

func (sc *SarpcClient) GetPaymentQueryInfo(encodedExtrinsic string) (paymentInfo *rpc.PaymentQueryInfo, err error) {
	v := &rpc.JsonRpcResult{}
	if err = sc.sendWsRequest(nil, v, rpc.SystemPaymentQueryInfo(wsId, encodedExtrinsic)); err != nil {
		return
	}

	paymentInfo = v.ToPaymentQueryInfo()
	if paymentInfo == nil {
		return nil, fmt.Errorf("get PaymentQueryInfo error")
	}
	return
}
