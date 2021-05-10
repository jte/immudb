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

package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/codenotary/immudb/pkg/pgsql/server/pgmeta"
	"net"
)

type rawMessage struct {
	t       byte
	payload []byte
}

type messageReader struct{}

type MessageReader interface {
	ReadRawMessage(conn net.Conn) (*rawMessage, error)
	WriteMessage(conn net.Conn, msg []byte) (int, error)
}

func NewMessageReader() *messageReader {
	return &messageReader{}
}

func (r *messageReader) ReadRawMessage(conn net.Conn) (*rawMessage, error) {
	t := make([]byte, 1)
	if _, err := conn.Read(t); err != nil {
		return nil, err
	}

	if _, ok := pgmeta.MTypes[t[0]]; !ok {
		return nil, errors.New(fmt.Sprintf(ErrUnknowMessageType.Error()+". Message first byte was %s", string(t[0])))
	}

	lb := make([]byte, 4)
	if _, err := conn.Read(lb); err != nil {
		return nil, err
	}
	l := binary.BigEndian.Uint32(lb)
	payload := make([]byte, l-4)
	if _, err := conn.Read(payload); err != nil {
		return nil, err
	}

	return &rawMessage{
		t:       t[0],
		payload: payload,
	}, nil
}

func (r *messageReader) WriteMessage(conn net.Conn, msg []byte) (int, error) {
	return conn.Write(msg)
}