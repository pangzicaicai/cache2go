/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package cache

import (
	"sync"
	"time"
	"log"
	"sort"
)

type Table struct {
	// The zero value for a RWMutex is an unlocked mutex.
	// 组合
	sync.RWMutex

	name            string
	// interface{} 指代任意类型
	items           map[interface{}]*Item
	cleanupTimer    *time.Timer
	cleanupInterval time.Duration

	logger *log.Logger

	// Callback method triggered when trying to load a non-existing key.
	loadData func(key interface{}, args ...interface{}) *Item
	// Callback method triggered when adding a new item from the cache.
	addedItem func(item *Item)
	// Callback method triggered before deleting an item from the cache.
	aboutToDeleteItem func(item *Item)
}

/**
	返回table中存储的items数量
 */
func (table *Table) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

/**
	遍历循环 cacheTable, 读出所有的items
	并将遍历出来的 键-值对传入 指定的方法中
 */
func (table *Table) Foreach(trans func(key interface{}, item *Item)) {
	table.RLock()
	defer table.Unlock()

	for k, v := range table.items {
		trans(k, v)
	}
}

/**
	当访问不存在的key的时候调用, 可以传递多个参数
	func(interface{}, ...interface{}) {}
	传入至少一个参数, ...interface{}, 代码可以传入任意类型的任意个参数
 */
func (table *Table) SetDataLoader(f func(interface{}, ...interface{}) *Item) {
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}

/**
	设置添加item的回调函数
 */
func (table *Table) SetAddedItemCallback(f func(*Item)) {
	table.Lock()
	defer table.Unlock()
	table.addedItem = f
}

/**
	设置要删除 item 的时候的回调函数
 */
func (table *Table) SetAboutToDeleteItemCallback(f func(*Item)) {
	table.Lock()
	defer table.Unlock()
	table.aboutToDeleteItem = f
}

/**
	设置logger
 */
func (table *Table) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()

	table.logger = logger
}

/*
	Expiration check loop, triggered by a self-adjusting timer.
	过期时间检测, 由自调整的 时间触发器触发
 */
func (table *Table) expirationCheck() {
	table.Lock()
	defer table.Unlock()
	// 定期清理定时器
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	// 触发时间间隔
	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after,",
			table.cleanupInterval,
			"for table",
			table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}

	// 当前时间 Local Time
	now := time.Now()
	// 最小时间间隔
	smallestDuration := 0 * time.Second

	for key, item := range table.items {
		item.RLock()
		lifeTime := item.LifeTime()

		accessedAt := item.accessedAt
		item.Unlock()

		if lifeTime == 0 {
			continue
		}

		// 这里的生命周期是指:
		// 从上一次访问这个key的时候开始算的
		if now.Sub(accessedAt) >= lifeTime {
			table.deleteInternal(key)
		} else {
			// 调整最小的时间间隔
			if smallestDuration == 0 ||
				lifeTime-now.Sub(accessedAt) < smallestDuration {
				smallestDuration = lifeTime - now.Sub(accessedAt)
			}
		}
	}

	// Setup the interval for the next cleanup run.
	// 设置下次定期清理的时间间隔
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			// 这里并不需要使用 go func()
			// 此处在timer的 goroutine 中调用
			table.expirationCheck()
		})
	}

}

func (table *Table) addInternal(item *Item) {

	table.log("Adding item with key,", item.key,
		"and lifetime of", item.lifeTime)

	table.items[item.key] = item

	expDuration := table.cleanupInterval
	addedItemCallback := table.addedItem

	table.Unlock()
	if addedItemCallback != nil {
		addedItemCallback(item)
	}

	if item.lifeTime > 0 && (expDuration == 0 || item.lifeTime < expDuration) {
		go table.expirationCheck()
	}
}

func (table *Table) Add(key interface{}, value interface{}, lifeTime time.Duration) *Item {
	item := NewItem(key, value, lifeTime)

	// Add item to cache.
	table.Lock()
	table.addInternal(item)

	return item
}

func (table *Table) deleteInternal(key interface{}) (*Item, error) {
	// map 访问不存在的key 并不会报错
	// 不存在是 ok 为 false
	item, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	aboutToDeleteItemCallback := table.aboutToDeleteItem
	table.Unlock()

	if aboutToDeleteItemCallback != nil {
		aboutToDeleteItemCallback(item)
	}

	item.RLock()
	defer item.RUnlock()

	if item.aboutToExpire != nil {
		item.aboutToExpire(key)
	}

	table.Lock()
	table.log("Deleting item with key", key,
		"created at", item.createdAt,
		"and hit", item.accessCount,
		"times from table", table.name)

	delete(table.items, key)

	return item, nil

}

func (table *Table) Delete(key interface{}) (*Item, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInternal(key)
}

func (table *Table) Exists(key interface{}) bool {
	table.RLock()
	defer table.Unlock()

	_, ok := table.items[key]

	return ok
}

/**
	找到这个key 或者 添加这个key
 */
func (table *Table) NotFoundAdd(key interface{}, value interface{}, lifeTime time.Duration) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := NewItem(key, value, lifeTime)
	table.addInternal(item)

	return true
}

func (table *Table) Value(key interface{}, args ...interface{}) (*Item, error) {
	table.RLock()
	item, ok := table.items[key]
	loadDataCallback := table.loadData
	table.Unlock()

	if ok {
		item.KeepAlive()
		return item, nil
	}
	// 参数类型 前面加 ... 代表 多个参数
	// 参数后面加 .. 代表打包穿进去
	if loadDataCallback != nil {
		item := loadDataCallback(key, args...)
		if item != nil {
			table.Add(key, item.value, item.lifeTime)
			return item, nil
		}

		return nil, ErrKeyNotFoundOrLoadable
	}

	return nil, ErrKeyNotFound
}

func (table *Table) Flush() {
	table.Lock()
	defer table.Unlock()

	table.log("Flushing table", table.name)

	table.items = make(map[interface{}]*Item)

	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}
func (table *Table) log(v ...interface{}) {
	if table.logger == nil {
		return
	}

	table.logger.Println(v)
}

// maps key to access counter
type ItemPair struct {
	Key         interface{}
	AccessCount int64
}

/**
	ItemPairList 是 []ItemPair的别名
	实现了排序接口,  Sorter
	排序次数按照访问次数来排序
 */
type ItemPairList []ItemPair

func (list ItemPairList) Len() int {
	return len(list)
}

// ???
func (list ItemPairList) Less(i, j int) bool {
	return list[i].AccessCount > list[j].AccessCount
}

func (list ItemPairList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

/**
	获取前 count 个 访问次数最多的item
 */
func (table *Table) MostAccessed(count int64) []*Item {
	table.RLock()
	defer table.RUnlock()

	p := make(ItemPairList, len(table.items))

	i := 0
	for k, v := range table.items {
		p[i] = ItemPair{k, v.accessCount}
		i ++
	}

	sort.Sort(p)

	var r []*Item
	c := int64(0)

	for _, v := range p {
		if c >= count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}

		c ++
	}

	return r
}
