package cluster

import (
	"github.com/hdt3213/godis/lib/utils"
	"github.com/hdt3213/godis/redis/connection"
	"github.com/hdt3213/godis/redis/parser"
	"github.com/hdt3213/godis/redis/reply/asserts"
	"testing"
)

func TestPublish(t *testing.T) {
	channel := utils.RandString(5)
	msg := utils.RandString(5)
	conn := &connection.FakeConn{}
	Subscribe(testCluster, conn, utils.ToBytesList("SUBSCRIBE", channel))
	conn.Clean() // clean subscribe success
	Publish(testCluster, conn, utils.ToBytesList("PUBLISH", channel, msg))
	data := conn.Bytes()
	ret, err := parser.ParseOne(data)
	if err != nil {
		t.Error(err)
		return
	}
	asserts.AssertMultiBulkReply(t, ret, []string{
		"message",
		channel,
		msg,
	})

	// unsubscribe
	UnSubscribe(testCluster, conn, utils.ToBytesList("UNSUBSCRIBE", channel))
	conn.Clean()
	Publish(testCluster, conn, utils.ToBytesList("PUBLISH", channel, msg))
	data = conn.Bytes()
	if len(data) > 0 {
		t.Error("expect no msg")
	}

	// unsubscribe all
	Subscribe(testCluster, conn, utils.ToBytesList("SUBSCRIBE", channel))
	UnSubscribe(testCluster, conn, utils.ToBytesList("UNSUBSCRIBE"))
	conn.Clean()
	Publish(testCluster, conn, utils.ToBytesList("PUBLISH", channel, msg))
	data = conn.Bytes()
	if len(data) > 0 {
		t.Error("expect no msg")
	}
}
