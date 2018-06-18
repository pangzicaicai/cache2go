/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */
package cache

import "errors"

var (
	ErrKeyNotFound = errors.New(
		"Key not found in cache.")
	ErrKeyNotFoundOrLoadable = errors.New(
		"Key not found and could not loaded into cache.")
)
