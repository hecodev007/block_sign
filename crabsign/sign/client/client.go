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

package client

import (
	"context"
	"log"

	"wallet-sign/sign/config"
	gethrpc "wallet-sign/sign/gethrpc"
	"wallet-sign/sign/types"
)

type Client interface {
	// Call makes the call to RPC method with the provided args
	// args must be encoded in the format RPC understands
	Call(result interface{}, method string, args ...interface{}) error

	Subscribe(ctx context.Context, namespace, subscribeMethodSuffix, unsubscribeMethodSuffix,
		notificationMethodSuffix string, channel interface{}, args ...interface{}) (
		*gethrpc.ClientSubscription, error)

	URL() string
}

type client struct {
	gethrpc.Client

	url string
}

// URL returns the URL the client connects to
func (c client) URL() string {
	return c.url
}

// Connect connects to the provided url
func Connect(url string) (Client, error) {
	log.Printf("Connecting to %v...", url)

	ctx, cancel := context.WithTimeout(context.Background(), config.Default().DialTimeout)
	defer cancel()

	c, err := gethrpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	cc := client{*c, url}
	return &cc, nil
}

func CallWithBlockHash(c Client, target interface{}, method string, blockHash *types.Hash, args ...interface{}) error {
	if blockHash == nil {
		err := c.Call(target, method, args...)
		if err != nil {
			return err
		}
		return nil
	}
	hexHash, err := types.Hex(*blockHash)
	if err != nil {
		return err
	}
	hargs := append(args, hexHash)
	err = c.Call(target, method, hargs...)
	if err != nil {
		return err
	}
	return nil
}
