/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2012, Radu Ioan Fericean
 *                   2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package cache

import "sync"

var (
	cache = make(map[string]*Table)
	mutex sync.RWMutex // 读写锁
)

func Cache(table string) *Table {
	// 读锁, 共享锁, 锁住的时候, 都可以读，但是不能写
	// 同时呢, 只有持有锁的地方可以加写锁, 其他地方不能加写锁
	mutex.RLock()
	// map, 如果key不存在, 那么 ok 为false
	t, ok := cache[table]
	mutex.RUnlock()

	if !ok {
		// 这里为什么还要再判断一次 ??
		mutex.Lock()
		t, ok = cache[table]
		if !ok {
			// 如果没有那么就创建一个
			t = &Table{
				name: table,
				// key 为 interface 表示 任何类型都可以,
				// 这里理解为 只要是可哈希的类型即可
				// 在Go里面, 可哈希的类型包括 ??
				items: make(map[interface{}]*Item),
			}

			cache[table] = t
		}
		mutex.Unlock()
	}
	
	return t
}
