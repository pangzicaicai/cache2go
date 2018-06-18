/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */


package cache

import (
	"time"
	"sync"
)

/**
	Item is a key-value cache item.
	with a sync.RWMutex to ensure atomic
	在cache包下, 因此写 Item就好, 无需写 ca
 */
type Item struct {
	/**
		RWMutex读写锁
		读锁(共享锁): 读的时候, 其他人也可以读, 但是不能写, 自己可以加写锁, 其他人不能加
		写锁(排它锁): 只能自己读写, 其他人不能读写
	 */
	sync.RWMutex

	/**
		在Go中, 使用interface{} 来指代有类型
	 */
	// The item's key
	key interface{}
	// The item's value
	value interface{}

	lifeTime time.Duration

	createdAt time.Time
	// Last assess timestamp.
	accessedAt time.Time
	// How often the time was accessed.
	accessCount int64

	// Callback method triggered right before removing the item from the cache
	aboutToExpire func(key interface{})
}

/**
	构造函数

 */
func NewItem(key interface{}, value interface{}, lifeTime time.Duration) *Item {
	t := time.Now() // 当前时间戳 returns the current local time.
	return &Item{
		key: key,
		value: value,
		createdAt: t,
		accessedAt: t,
		accessCount: 0,
		aboutToExpire: nil,
		lifeTime: lifeTime,
	}
}

func (item *Item) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedAt = time.Now()
	item.accessCount ++
}
/**
	Return this item's expiration duration.
 */
func (item *Item) LifeTime() time.Duration {
	return item.lifeTime
}

/**
	Return the last access time
 */
func (item *Item) AccessedAt() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.AccessedAt()
}

func (item *Item) CreatedAt() time.Time {
	return item.createdAt
}

func (item *Item) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

func (item *Item) Key() interface{} {
	return item.key
}

func (item *Item) Value() interface{} {
	return item.value
}

func (item *Item) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = f
}