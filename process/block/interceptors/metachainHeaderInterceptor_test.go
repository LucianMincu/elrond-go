package interceptors_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewMetachainHeaderInterceptor

func TestNewMetachainHeaderInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		nil,
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainHeadersShouldErr(t *testing.T) {
	t.Parallel()

	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMetaHeadersDataPool, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		&mock.CacherStub{},
		nil,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMetaHeadersNoncesDataPool, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetaHeaderValidatorShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		nil,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHeaderHandlerValidator, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Nil(t, mhi)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewMetachainHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		nil,
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		nil,
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilNodesCoordinator, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, mhi)
}

//------- ProcessReceivedMessage

func TestMetachainHeaderInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMessage, mhi.ProcessReceivedMessage(nil))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")
	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}
	multisigner := mock.NewMultiSigner()
	nodesCoordinator := mock.NewNodesCoordinatorMock()

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		headerValidator,
		multisigner,
		hasher,
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedMetaHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilPubKeysBitmap, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}
	chanDone := make(chan struct{}, 1)
	testedNonce := uint64(67)
	metachainHeaders := &mock.CacherStub{}
	metachainHeadersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
	}
	multisigner := mock.NewMultiSigner()
	nodesCoordinator := &mock.NodesCoordinatorMock{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainHeadersNonces,
		headerValidator,
		multisigner,
		hasher,
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedMetaHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	hdr.Nonce = testedNonce
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = []byte{1, 0, 0}
	hdr.Signature = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	metachainHeaders.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wg.Done()
		}
		return
	}
	metachainHeadersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce != testedNonce {
			return
		}

		aaaHash := mock.HasherMock{}.Compute(string(buff))
		src.Range(func(sharId uint32, hash []byte) bool {
			if bytes.Equal(aaaHash, hash) {
				wg.Done()

				return false
			}

			return true
		})
	}

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	assert.Nil(t, mhi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageIsNotValidShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}
	chanDone := make(chan struct{}, 1)
	testedNonce := uint64(67)
	multisigner := mock.NewMultiSigner()
	metachainHeaders := &mock.CacherStub{}
	metachainHeadersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return false
		},
	}

	nodesCoordinator := &mock.NodesCoordinatorMock{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainHeadersNonces,
		headerValidator,
		multisigner,
		hasher,
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedMetaHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	hdr.Nonce = testedNonce
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = []byte{1, 0, 0}
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	metachainHeaders.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			chanDone <- struct{}{}
		}
		return
	}
	metachainHeadersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce != testedNonce {
			return
		}

		aaaHash := mock.HasherMock{}.Compute(string(buff))
		src.Range(func(shardId uint32, hash []byte) bool {
			if bytes.Equal(aaaHash, hash) {
				chanDone <- struct{}{}

				return false
			}

			return true
		})
	}

	assert.Nil(t, mhi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add block in pool")
	case <-time.After(durTimeout):
	}
}
