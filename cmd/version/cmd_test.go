/*
Copyright 2019-2020 vChain, Inc.

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

package version

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	cmd := VersionCmd()

	// no version info
	collector := new(StdOutCollector)
	require.NoError(t, collector.Start())
	require.NoError(t, cmd.Execute())
	noVersionOutput, err := collector.Stop()
	require.NoError(t, err)
	require.Equal(t, "no version info available\n", noVersionOutput)

	// full version info
	App = "Some App"
	Version = "v1.0.0"
	Commit = "2F20B9ADF24C82A6AFEE0CEBF53B46A512FE9526"
	BuiltBy = "some.user@somedomain.com"
	builtAt, _ := time.Parse(time.RFC3339, "2020-07-13T23:28:09Z")
	BuiltAt = fmt.Sprintf("%d", builtAt.Unix())
	Static = "static"

	builtAtUnix, _ := strconv.ParseInt(BuiltAt, 10, 64)
	builtAtStr := time.Unix(builtAtUnix, 0).Format(time.RFC1123)
	expectedVersionOutput := strings.Join(
		[]string{
			"Some App v1.0.0",
			"Commit  : 2F20B9ADF24C82A6AFEE0CEBF53B46A512FE9526",
			"Built by: some.user@somedomain.com",
			"Built at: " + builtAtStr,
			"Static  : true\n",
		},
		"\n")
	require.NoError(t, collector.Start())
	require.NoError(t, cmd.Execute())
	versionOutput, err := collector.Stop()
	require.NoError(t, err)
	require.Equal(t, expectedVersionOutput, versionOutput)
}

type StdOutCollector struct {
	realStdOut       *os.File
	fakeStdOutReader *os.File
	fakeStdOutWriter *os.File
}

func (c *StdOutCollector) Start() error {
	c.realStdOut = os.Stdout // keep backup of the real stdout
	var err error
	c.fakeStdOutReader, c.fakeStdOutWriter, err = os.Pipe()
	if err != nil {
		return err
	}
	os.Stdout = c.fakeStdOutWriter
	return nil
}

func (c *StdOutCollector) Stop() (string, error) {
	outC := make(chan string)
	outErr := make(chan error)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, c.fakeStdOutReader)
		if err != nil {
			outErr <- err
		}
		outC <- buf.String()
	}()

	// back to normal state
	c.fakeStdOutWriter.Close()
	os.Stdout = c.realStdOut // restore the real stdout
	select {
	case out := <-outC:
		return out, nil
	case err := <-outErr:
		return "", err
	}
}