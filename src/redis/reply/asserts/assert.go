package asserts

import (
	"fmt"
	"github.com/HDT3213/godis/src/datastruct/utils"
	"github.com/HDT3213/godis/src/interface/redis"
	"github.com/HDT3213/godis/src/redis/reply"
	"runtime"
	"testing"
)

func AssertIntReply(t *testing.T, actual redis.Reply, expected int) {
	intResult, ok := actual.(*reply.IntReply)
	if !ok {
		t.Errorf("expected int reply, actually %s, %s", actual.ToBytes(), printStack())
		return
	}
	if intResult.Code != int64(expected) {
		t.Errorf("expected %d, actually %d, %s", expected, intResult.Code, printStack())
	}
}

func AssertBulkReply(t *testing.T, actual redis.Reply, expected string) {
	bulkReply, ok := actual.(*reply.BulkReply)
	if !ok {
		t.Errorf("expected bulk reply, actually %s, %s", actual.ToBytes(), printStack())
		return
	}
	if !utils.BytesEquals(bulkReply.Arg, []byte(expected)) {
		t.Errorf("expected %s, actually %s, %s", expected, actual.ToBytes(), printStack())
	}
}

func AssertStatusReply(t *testing.T, actual redis.Reply, expected string) {
	statusReply, ok := actual.(*reply.StatusReply)
	if !ok {
		t.Errorf("expected bulk reply, actually %s, %s", actual.ToBytes(), printStack())
		return
	}
	if statusReply.Status != expected {
		t.Errorf("expected %s, actually %s, %s", expected, actual.ToBytes(), printStack())
	}
}

func AssertMultiBulkReply(t *testing.T, actual redis.Reply, expected []string) {
	multiBulk, ok := actual.(*reply.MultiBulkReply)
	if !ok {
		t.Errorf("expected bulk reply, actually %s, %s", actual.ToBytes(), printStack())
		return
	}
	if len(multiBulk.Args) != len(expected) {
		t.Errorf("expected %d elements, actually %d, %s",
			len(expected), len(multiBulk.Args), printStack())
		return
	}
	for i, v := range multiBulk.Args {
		str := string(v)
		if str != expected[i] {
			t.Errorf("expected %s, actually %s, %s", expected[i], actual, printStack())
		}
	}
}

func AssertMultiBulkReplySize(t *testing.T, actual redis.Reply, expected int) {
	multiBulk, ok := actual.(*reply.MultiBulkReply)
	if !ok {
		t.Errorf("expected bulk reply, actually %s, %s", actual.ToBytes(), printStack())
		return
	}
	if len(multiBulk.Args) != expected {
		t.Errorf("expected %d elements, actually %d, %s", expected, len(multiBulk.Args), printStack())
		return
	}
}

func printStack() string {
	_, file, no, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("at %s#%d", file, no)
	}
	return ""
}
