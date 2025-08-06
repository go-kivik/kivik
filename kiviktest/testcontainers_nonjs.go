// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build !js

package kiviktest

import (
	"context"
	"os"
	"testing"

	tc "github.com/go-kivik/kivik/v4/kiviktest/testcontainers"
)

func startCouchDB(t *testing.T, image string) string { //nolint:thelper // Not a helper
	if os.Getenv("USETC") == "" {
		t.Skip("USETC not set, skipping testcontainers")
	}

	dsn, err := tc.StartCouchDB(tContext(t), image)
	if err != nil {
		t.Fatal(err)
	}
	return dsn
}

func tContext(t *testing.T) context.Context {
	t.Helper()
	if c, ok := interface{}(t).(interface{ Context() context.Context }); ok {
		return c.Context()
	}
	return context.Background()
}
