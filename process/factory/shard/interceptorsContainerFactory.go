package shard

import (
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const maxGoRoutineTxInterceptor = 100

type interceptorsContainerFactory struct {
	accounts               state.AccountsAdapter
	shardCoordinator       sharding.Coordinator
	messenger              process.TopicHandler
	store                  dataRetriever.StorageService
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	keyGen                 crypto.KeyGenerator
	singleSigner           crypto.SingleSigner
	multiSigner            crypto.MultiSigner
	dataPool               dataRetriever.PoolsHolder
	addrConverter          state.AddressConverter
	nodesCoordinator       sharding.NodesCoordinator
	txInterceptorThrottler process.InterceptorThrottler
	maxTxNonceDeltaAllowed int
	txFeeHandler           process.FeeHandler
}

// NewInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewInterceptorsContainerFactory(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	messenger process.TopicHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	singleSigner crypto.SingleSigner,
	multiSigner crypto.MultiSigner,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
	maxTxNonceDeltaAllowed int,
	txFeeHandler process.FeeHandler,
) (*interceptorsContainerFactory, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilBlockChain
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if keyGen == nil || keyGen.IsInterfaceNil() {
		return nil, process.ErrNilKeyGen
	}
	if singleSigner == nil || singleSigner.IsInterfaceNil() {
		return nil, process.ErrNilSingleSigner
	}
	if multiSigner == nil || multiSigner.IsInterfaceNil() {
		return nil, process.ErrNilMultiSigVerifier
	}
	if dataPool == nil || dataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if nodesCoordinator == nil || nodesCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilNodesCoordinator
	}
	if txFeeHandler == nil || txFeeHandler.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	txInterceptorThrottler, err := throttler.NewNumGoRoutineThrottler(maxGoRoutineTxInterceptor)
	if err != nil {
		return nil, err
	}

	return &interceptorsContainerFactory{
		accounts:               accounts,
		shardCoordinator:       shardCoordinator,
		nodesCoordinator:       nodesCoordinator,
		messenger:              messenger,
		store:                  store,
		marshalizer:            marshalizer,
		hasher:                 hasher,
		keyGen:                 keyGen,
		singleSigner:           singleSigner,
		multiSigner:            multiSigner,
		dataPool:               dataPool,
		addrConverter:          addrConverter,
		txInterceptorThrottler: txInterceptorThrottler,
		maxTxNonceDeltaAllowed: maxTxNonceDeltaAllowed,
		txFeeHandler:           txFeeHandler,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (icf *interceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	keys, interceptorSlice, err := icf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateRewardTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateHdrInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generatePeerChBlockBodyInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateMetachainHeaderInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (icf *interceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
) (process.Interceptor, error) {

	err := icf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return interceptor, icf.messenger.RegisterMessageProcessor(topic, interceptor)
}

//------- Tx interceptors

func (icf *interceptorsContainerFactory) generateTxInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierTx
		interceptorSlice[int(idx)] = interceptor
	}

	//tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)
	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneTxInterceptor(identifier string) (process.Interceptor, error) {
	txValidator, err := dataValidators.NewTxValidator(icf.accounts, icf.shardCoordinator, icf.maxTxNonceDeltaAllowed)
	if err != nil {
		return nil, err
	}

	interceptor, err := transaction.NewTxInterceptor(
		icf.marshalizer,
		icf.dataPool.Transactions(),
		txValidator,
		icf.addrConverter,
		icf.hasher,
		icf.singleSigner,
		icf.keyGen,
		icf.shardCoordinator,
		icf.txInterceptorThrottler,
		icf.txFeeHandler,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Reward transactions interceptors

func (icf *interceptorsContainerFactory) generateRewardTxInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.RewardsTransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneRewardTxInterceptor(identifierScr)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierScr
		interceptorSlice[int(idx)] = interceptor
	}

	identifierTx := factory.RewardsTransactionTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneRewardTxInterceptor(identifierTx)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)

	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneRewardTxInterceptor(identifier string) (process.Interceptor, error) {
	rewardTxStorer := icf.store.GetStorer(dataRetriever.RewardTransactionUnit)

	interceptor, err := rewardTransaction.NewRewardTxInterceptor(
		icf.marshalizer,
		icf.dataPool.RewardTransactions(),
		rewardTxStorer,
		icf.addrConverter,
		icf.hasher,
		icf.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Unsigned transactions interceptors

func (icf *interceptorsContainerFactory) generateUnsignedTxsInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierScr
		interceptorSlice[int(idx)] = interceptor
	}

	identifierTx := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneUnsignedTxInterceptor(identifierTx)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)
	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneUnsignedTxInterceptor(identifier string) (process.Interceptor, error) {
	uTxStorer := icf.store.GetStorer(dataRetriever.UnsignedTransactionUnit)

	interceptor, err := unsigned.NewUnsignedTxInterceptor(
		icf.marshalizer,
		icf.dataPool.UnsignedTransactions(),
		uTxStorer,
		icf.addrConverter,
		icf.hasher,
		icf.shardCoordinator)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Hdr interceptor

func (icf *interceptorsContainerFactory) generateHdrInterceptor() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, nil, err
	}

	//only one intrashard header topic
	identifierHdr := factory.HeadersTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	interceptor, err := interceptors.NewHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.Headers(),
		icf.dataPool.HeadersNonces(),
		hdrValidator,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.nodesCoordinator,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}

//------- MiniBlocks interceptors

func (icf *interceptorsContainerFactory) generateMiniBlocksInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+1)
	interceptorsSlice := make([]process.Interceptor, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierMiniBlocks
		interceptorsSlice[int(idx)] = interceptor
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
	if err != nil {
		return nil, nil, err
	}

	keys[noOfShards] = identifierMiniBlocks
	interceptorsSlice[noOfShards] = interceptor

	return keys, interceptorsSlice, nil
}

func (icf *interceptorsContainerFactory) createOneMiniBlocksInterceptor(identifier string) (process.Interceptor, error) {
	txBlockBodyStorer := icf.store.GetStorer(dataRetriever.MiniBlockUnit)

	interceptor, err := interceptors.NewTxBlockBodyInterceptor(
		icf.marshalizer,
		icf.dataPool.MiniBlocks(),
		txBlockBodyStorer,
		icf.hasher,
		icf.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- PeerChBlocks interceptor

func (icf *interceptorsContainerFactory) generatePeerChBlockBodyInterceptor() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := factory.PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	peerBlockBodyStorer := icf.store.GetStorer(dataRetriever.PeerChangesUnit)

	interceptor, err := interceptors.NewPeerBlockBodyInterceptor(
		icf.marshalizer,
		icf.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		icf.hasher,
		shardC,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierPeerCh, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierPeerCh}, []process.Interceptor{interceptor}, nil
}

//------- MetachainHeader interceptors

func (icf *interceptorsContainerFactory) generateMetachainHeaderInterceptor() ([]string, []process.Interceptor, error) {
	identifierHdr := factory.MetachainBlocksTopic
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, nil, err
	}

	interceptor, err := interceptors.NewMetachainHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.MetaBlocks(),
		icf.dataPool.HeadersNonces(),
		hdrValidator,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.nodesCoordinator,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (icf *interceptorsContainerFactory) IsInterfaceNil() bool {
	if icf == nil {
		return true
	}
	return false
}
