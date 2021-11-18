// +build streams
/*
Copyright 2021 CodeNotary, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"context"
	ic "github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"os"
	"testing"
)

func TestTransaction_SetAndGet(t *testing.T) {
	options := server.DefaultOptions()
	bs := servertest.NewBufconnServer(options)

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	bs.Start()
	defer bs.Stop()

	client := ic.DefaultClient().WithOptions(ic.DefaultOptions().WithDialOptions([]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()}))

	serverUUID, sessionID, err := client.OpenSession(context.TODO(), []byte(`immudb`), []byte(`immudb`), "defaultdb")
	require.NoError(t, err)
	require.NotNil(t, serverUUID)
	require.NotNil(t, sessionID)

	tx, err := client.BeginTx(context.TODO(), &ic.TxOptions{ReadWrite: true})
	require.NoError(t, err)
	err = tx.Set(context.TODO(), []byte(`key`), []byte(`val`))
	require.NoError(t, err)
	kv, err := tx.Get(context.TODO(), []byte(`key`))
	require.NoError(t, err)
	require.Equal(t, []byte(`key`), kv.Key)
	require.Equal(t, []byte(`val`), kv.Value)
	tx.Commit(context.TODO())

	err = client.CloseSession(context.TODO())
	require.NoError(t, err)
}
