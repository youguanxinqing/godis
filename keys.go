package godis

import (
	"github.com/hdt3213/godis/datastruct/dict"
	"github.com/hdt3213/godis/datastruct/list"
	"github.com/hdt3213/godis/datastruct/set"
	"github.com/hdt3213/godis/datastruct/sortedset"
	"github.com/hdt3213/godis/interface/redis"
	"github.com/hdt3213/godis/lib/wildcard"
	"github.com/hdt3213/godis/redis/reply"
	"strconv"
	"time"
)

// execDel removes a key from db
func execDel(db *DB, args [][]byte) redis.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}

	db.Locks(keys...)
	defer db.UnLocks(keys...)

	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.AddAof(makeAofCmd("del", args))
	}
	return reply.MakeIntReply(int64(deleted))
}

// execExists checks if a is existed in db
func execExists(db *DB, args [][]byte) redis.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// execFlushDB removes all data in current db
func execFlushDB(db *DB, args [][]byte) redis.Reply {
	db.Flush()
	db.AddAof(makeAofCmd("flushdb", args))
	return &reply.OkReply{}
}

// execFlushAll removes all data in all db
func execFlushAll(db *DB, args [][]byte) redis.Reply {
	db.Flush()
	db.AddAof(makeAofCmd("flushdb", args))
	return &reply.OkReply{}
}

// execType returns the type of entity, including: string, list, hash, set and zset
func execType(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	case *list.LinkedList:
		return reply.MakeStatusReply("list")
	case dict.Dict:
		return reply.MakeStatusReply("hash")
	case *set.Set:
		return reply.MakeStatusReply("set")
	case *sortedset.SortedSet:
		return reply.MakeStatusReply("zset")
	}
	return &reply.UnknownErrReply{}
}

// execRename a key
func execRename(db *DB, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[0])
	dest := string(args[1])

	db.Locks(src, dest)
	defer db.UnLocks(src, dest)

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	rawTTL, hasTTL := db.ttlMap.Get(src)
	db.PutEntity(dest, entity)
	db.Remove(src)
	if hasTTL {
		db.Persist(src) // clean src and dest with their ttl
		db.Persist(dest)
		expireTime, _ := rawTTL.(time.Time)
		db.Expire(dest, expireTime)
	}
	db.AddAof(makeAofCmd("rename", args))
	return &reply.OkReply{}
}

// execRenameNx a key, only if the new key does not exist
func execRenameNx(db *DB, args [][]byte) redis.Reply {
	src := string(args[0])
	dest := string(args[1])

	db.Locks(src, dest)
	defer db.UnLocks(src, dest)

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	rawTTL, hasTTL := db.ttlMap.Get(src)
	db.Removes(src, dest) // clean src and dest with their ttl
	db.PutEntity(dest, entity)
	if hasTTL {
		db.Persist(src) // clean src and dest with their ttl
		db.Persist(dest)
		expireTime, _ := rawTTL.(time.Time)
		db.Expire(dest, expireTime)
	}
	db.AddAof(makeAofCmd("renamenx", args))
	return reply.MakeIntReply(1)
}

// execExpire sets a key's time to live in seconds
func execExpire(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])

	ttlArg, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	ttl := time.Duration(ttlArg) * time.Second

	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	expireAt := time.Now().Add(ttl)
	db.Expire(key, expireAt)
	db.AddAof(makeExpireCmd(key, expireAt))
	return reply.MakeIntReply(1)
}

// execExpireAt sets a key's expiration in unix timestamp
func execExpireAt(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])

	raw, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	expireTime := time.Unix(raw, 0)

	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	db.Expire(key, expireTime)
	db.AddAof(makeExpireCmd(key, expireTime))
	return reply.MakeIntReply(1)
}

// execPExpire sets a key's time to live in milliseconds
func execPExpire(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])

	ttlArg, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	ttl := time.Duration(ttlArg) * time.Millisecond

	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	expireTime := time.Now().Add(ttl)
	db.Expire(key, expireTime)
	db.AddAof(makeExpireCmd(key, expireTime))
	return reply.MakeIntReply(1)
}

// execPExpireAt sets a key's expiration in unix timestamp specified in milliseconds
func execPExpireAt(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])

	raw, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	expireTime := time.Unix(0, raw*int64(time.Millisecond))

	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	db.Expire(key, expireTime)

	db.AddAof(makeExpireCmd(key, expireTime))
	return reply.MakeIntReply(1)
}

// execTTL returns a key's time to live in seconds
func execTTL(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(-2)
	}

	raw, exists := db.ttlMap.Get(key)
	if !exists {
		return reply.MakeIntReply(-1)
	}
	expireTime, _ := raw.(time.Time)
	ttl := expireTime.Sub(time.Now())
	return reply.MakeIntReply(int64(ttl / time.Second))
}

// execPTTL returns a key's time to live in milliseconds
func execPTTL(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(-2)
	}

	raw, exists := db.ttlMap.Get(key)
	if !exists {
		return reply.MakeIntReply(-1)
	}
	expireTime, _ := raw.(time.Time)
	ttl := expireTime.Sub(time.Now())
	return reply.MakeIntReply(int64(ttl / time.Millisecond))
}

// execPersist removes expiration from a key
func execPersist(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	_, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	_, exists = db.ttlMap.Get(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	db.Persist(key)
	db.AddAof(makeAofCmd("persist", args))
	return reply.MakeIntReply(1)
}

// BGRewriteAOF asynchronously rewrites Append-Only-File
func BGRewriteAOF(db *DB, args [][]byte) redis.Reply {
	go db.aofRewrite()
	return reply.MakeStatusReply("Background append only file rewriting started")
}

// execKeys returns all keys matching the given pattern
func execKeys(db *DB, args [][]byte) redis.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("Del", execDel, nil, -2)
	RegisterCommand("Expire", execExpire, nil, 3)
	RegisterCommand("ExpireAt", execExpireAt, nil, 3)
	RegisterCommand("PExpire", execPExpire, nil, 3)
	RegisterCommand("PExpireAt", execPExpireAt, nil, 3)
	RegisterCommand("TTL", execTTL, nil, 2)
	RegisterCommand("PTTL", execPTTL, nil, 2)
	RegisterCommand("Persist", execPersist, nil, 2)
	RegisterCommand("Exists", execExists, nil, -2)
	RegisterCommand("Type", execType, nil, 2)
	RegisterCommand("Rename", execRename, nil, 3)
	RegisterCommand("RenameNx", execRenameNx, nil, 3)
	RegisterCommand("FlushDB", execFlushDB, nil, -1)
	RegisterCommand("FlushAll", execFlushAll, nil, -1)
	RegisterCommand("Keys", execKeys, nil, 2)
}
