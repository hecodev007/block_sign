// Go Substrate RPC Client (GSRPC) provides APIs and types around Polkadot and any Substrate-based chain RPC calls
//
// Copyright 2019 Centrifuge GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"context"
	"sync"

	"github.com/group-coldwallet/wallet-sign/sign/config"
	gethrpc "github.com/group-coldwallet/wallet-sign/sign/gethrpc"
	"github.com/group-coldwallet/wallet-sign/sign/types"
)

// StorageSubscription is a subscription established through one of the Client's subscribe methods.
type StorageSubscription struct {
	sub      *gethrpc.ClientSubscription
	channel  chan types.StorageChangeSet
	quitOnce sync.Once // ensures quit is closed once
}

// Chan returns the subscription channel.
//
// The channel is closed when Unsubscribe is called on the subscription.
func (s *StorageSubscription) Chan() <-chan types.StorageChangeSet {
	return s.channel
}

// Err returns the subscription error channel. The intended use of Err is to schedule
// resubscription when the client connection is closed unexpectedly.
//
// The error channel receives a value when the subscription has ended due
// to an error. The received error is nil if Close has been called
// on the underlying client and no other error has occurred.
//
// The error channel is closed when Unsubscribe is called on the subscription.
func (s *StorageSubscription) Err() <-chan error {
	return s.sub.Err()
}

// Unsubscribe unsubscribes the notification and closes the error channel.
// It can safely be called more than once.
func (s *StorageSubscription) Unsubscribe() {
	s.sub.Unsubscribe()
	s.quitOnce.Do(func() {
		close(s.channel)
	})
}

// SubscribeStorageRaw subscribes the storage for the given keys, returning a subscription that will
// receive server notifications containing the storage change sets.
//
// Slow subscribers will be dropped eventually. Client buffers up to 20000 notifications before considering the
// subscriber dead. The subscription Err channel will receive ErrSubscriptionQueueOverflow. Use a sufficiently
// large buffer on the channel or ensure that the channel usually has at least one reader to prevent this issue.
func (s *State) SubscribeStorageRaw(keys []types.StorageKey) (
	*StorageSubscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Default().SubscribeTimeout)
	defer cancel()

	c := make(chan types.StorageChangeSet)

	keyss := make([]string, len(keys))
	for i := range keys {
		keyss[i] = keys[i].Hex()
	}

	sub, err := s.client.Subscribe(ctx, "state", "subscribeStorage", "unsubscribeStorage", "storage", c, keyss)
	if err != nil {
		return nil, err
	}

	return &StorageSubscription{sub: sub, channel: c}, nil
}
