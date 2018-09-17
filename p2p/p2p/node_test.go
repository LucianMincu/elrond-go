package p2p

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

var counter1 int32
var counter2 int32
var counter3 int32

func TestRecreationSameNode(t *testing.T) {

	port := 4000

	node1, err := CreateNewNode(context.Background(), port, []string{}, service.GetMarshalizerService())
	assert.Nil(t, err)

	node2, err := CreateNewNode(context.Background(), port, []string{}, service.GetMarshalizerService())
	assert.Nil(t, err)

	if node1.P2pNode.ID().Pretty() != node2.P2pNode.ID().Pretty() {
		t.Fatal("ID mismatch")
	}
}

func TestSimpleSend2NodesPingPong(t *testing.T) {
	node1, err := CreateNewNode(context.Background(), 5000, []string{}, service.GetMarshalizerService())
	assert.Nil(t, err)

	fmt.Println(node1.P2pNode.Addrs()[0].String())

	node2, err := CreateNewNode(context.Background(), 5001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()},
		service.GetMarshalizerService())
	assert.Nil(t, err)

	time.Sleep(time.Second)

	var val int32 = 0

	node1.OnMsgRecv = func(sender *Node, peerID string, m *Message) {
		fmt.Printf("Got message from peerID %v: %v\n", peerID, string(m.Payload))

		if string(m.Payload) == "Ping" {
			atomic.AddInt32(&val, 1)
			sender.SendDirectString(peerID, "Pong")
		}
	}

	node2.OnMsgRecv = func(sender *Node, peerID string, m *Message) {
		fmt.Printf("Got message from peerID %v: %v\n", peerID, string(m.Payload))

		if string(m.Payload) == "Ping" {
			sender.SendDirectString(peerID, "Pong")
		}

		if string(m.Payload) == "Pong" {
			atomic.AddInt32(&val, 1)
		}

	}

	node2.SendDirectString(node1.P2pNode.ID().Pretty(), "Ping")

	time.Sleep(time.Second)

	if val != 2 {
		t.Fatal("Should have been 2 (ping-pong pair)")
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()

}

func recv1(sender *Node, peerID string, m *Message) {
	atomic.AddInt32(&counter1, 1)
	fmt.Printf("%v > %v: Got message from peerID %v: %v\n", time.Now(), sender.P2pNode.ID().Pretty(), peerID, string(m.Payload))
	sender.BroadcastBuff(m.Payload, []string{peerID})
}

func recv2(sender *Node, peerID string, m *Message) {
	atomic.AddInt32(&counter2, 1)
	fmt.Printf("%v > %v: Got message from peerID %v: %v\n", time.Now(), sender.P2pNode.ID().Pretty(), peerID, string(m.Payload))
	sender.BroadcastString(string(m.Payload), []string{peerID})
}

func TestSimpleBroadcast5nodesInline(t *testing.T) {
	counter1 = 0

	marsh := service.GetMarshalizerService()

	node1, err := CreateNewNode(context.Background(), 6000, []string{}, marsh)
	assert.Nil(t, err)

	node2, err := CreateNewNode(context.Background(), 6001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	node3, err := CreateNewNode(context.Background(), 6002, []string{node2.P2pNode.Addrs()[0].String() + "/ipfs/" + node2.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	node4, err := CreateNewNode(context.Background(), 6003, []string{node3.P2pNode.Addrs()[0].String() + "/ipfs/" + node3.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	node5, err := CreateNewNode(context.Background(), 6004, []string{node4.P2pNode.Addrs()[0].String() + "/ipfs/" + node4.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	node1.OnMsgRecv = recv1
	node2.OnMsgRecv = recv1
	node3.OnMsgRecv = recv1
	node4.OnMsgRecv = recv1
	node5.OnMsgRecv = recv1

	fmt.Println()
	fmt.Println()

	node1.BroadcastString("Boo", []string{})

	time.Sleep(time.Second)

	if counter1 != 4 {
		t.Fatal("Should have been 4 (traversed 4 peers)")
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()
	node3.P2pNode.Close()
	node4.P2pNode.Close()
	node5.P2pNode.Close()

}

func TestSimpleBroadcast5nodesBeterConnected(t *testing.T) {
	counter2 = 0

	marsh := service.GetMarshalizerService()

	node1, err := CreateNewNode(context.Background(), 7000, []string{}, marsh)
	assert.Nil(t, err)

	node2, err := CreateNewNode(context.Background(), 7001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	node3, err := CreateNewNode(context.Background(), 7002, []string{node2.P2pNode.Addrs()[0].String() + "/ipfs/" + node2.P2pNode.ID().Pretty(),
		node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	node4, err := CreateNewNode(context.Background(), 7003, []string{node3.P2pNode.Addrs()[0].String() + "/ipfs/" + node3.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	node5, err := CreateNewNode(context.Background(), 7004, []string{node4.P2pNode.Addrs()[0].String() + "/ipfs/" + node4.P2pNode.ID().Pretty(),
		node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	node1.OnMsgRecv = recv2
	node2.OnMsgRecv = recv2
	node3.OnMsgRecv = recv2
	node4.OnMsgRecv = recv2
	node5.OnMsgRecv = recv2

	fmt.Println()
	fmt.Println()

	node1.BroadcastString("Boo", []string{})

	time.Sleep(time.Second)

	if counter2 != 4 {
		t.Fatal("Should have been 4 (traversed 4 peers), got", counter2)
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()
	node3.P2pNode.Close()
	node4.P2pNode.Close()
	node5.P2pNode.Close()
}

func TestMessageHops(t *testing.T) {
	marsh := service.GetMarshalizerService()

	node1, err := CreateNewNode(context.Background(), 8000, []string{}, marsh)
	assert.Nil(t, err)

	node2, err := CreateNewNode(context.Background(), 8001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()}, marsh)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	m := NewMessage(node1.P2pNode.ID().Pretty(), []byte("A"), marsh)

	var recv *Message = nil

	counter3 = 0

	node1.OnMsgRecv = func(sender *Node, peerID string, m *Message) {

		if counter3 < 10 {
			atomic.AddInt32(&counter3, 1)

			fmt.Printf("Node 1, recv %v, resending...\n", *m)
			m.AddHop(sender.P2pNode.ID().Pretty())
			sender.BroadcastMessage(m, []string{})

			recv = m
		}
	}

	node2.OnMsgRecv = func(sender *Node, peerID string, m *Message) {

		if counter3 < 10 {
			atomic.AddInt32(&counter3, 1)

			fmt.Printf("Node 2, recv %v, resending...\n", *m)
			m.AddHop(sender.P2pNode.ID().Pretty())
			sender.BroadcastMessage(m, []string{})

			recv = m
		}
	}

	node1.BroadcastMessage(m, []string{})

	time.Sleep(time.Second)

	if counter3 != 2 {
		t.Fatal(fmt.Sprintf("Shuld have been 2 iterations (messageQueue filtering), got %v!", counter3))
	}

	if recv == nil {
		t.Fatal("Not broadcasted?")
	}

	if recv.Hops != 2 {
		t.Fatal("Hops should have been 2")
	}

	if recv.Peers[0] != node1.P2pNode.ID().Pretty() {
		t.Fatal("hop 1 sould have been node's 1")
	}

	if recv.Peers[1] != node2.P2pNode.ID().Pretty() {
		t.Fatal("hop 2 sould have been node's 2")
	}

	if recv.Peers[2] != node1.P2pNode.ID().Pretty() {
		t.Fatal("hope 3 sould have been node's 1")
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()

}

func TestSendingNilShouldReturnError(t *testing.T) {
	node1, err := CreateNewNode(context.Background(), 9000, []string{}, service.GetMarshalizerService())
	assert.Nil(t, err)

	err = node1.BroadcastMessage(nil, []string{})
	assert.NotNil(t, err)

	err = node1.SendDirectMessage("", nil)
	assert.NotNil(t, err)

}

func TestMultipleErrorsOnBroadcasting(t *testing.T) {
	node1, err := CreateNewNode(context.Background(), 10000, []string{}, service.GetMarshalizerService())
	assert.Nil(t, err)

	node1.P2pNode.Peerstore().AddAddr("A", node1.P2pNode.Addrs()[0], peerstore.PermanentAddrTTL)

	err = node1.BroadcastString("aaa", []string{})
	assert.NotNil(t, err)

	if len(err.(*NodeError).NestedErrors) != 0 {
		t.Fatal("Should have had 0 nested errs")
	}

	node1.P2pNode.Peerstore().AddAddr("B", node1.P2pNode.Addrs()[0], peerstore.PermanentAddrTTL)

	err = node1.BroadcastString("aaa", []string{})
	assert.NotNil(t, err)

	if len(err.(*NodeError).NestedErrors) != 2 {
		t.Fatal("Should have had 2 nested errs")
	}

}

func TestCreateNodeWithNilMarshalizer(t *testing.T) {
	_, err := CreateNewNode(context.Background(), 11000, []string{}, nil)
	assert.NotNil(t, err)
}