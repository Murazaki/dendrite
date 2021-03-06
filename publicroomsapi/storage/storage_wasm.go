// Copyright 2020 The Matrix.org Foundation C.I.C.
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

package storage

import (
	"fmt"
	"net/url"

	"github.com/matrix-org/dendrite/publicroomsapi/storage/sqlite3"
	"github.com/matrix-org/gomatrixserverlib"
)

// NewPublicRoomsServerDatabase opens a database connection.
func NewPublicRoomsServerDatabase(dataSourceName string, localServerName gomatrixserverlib.ServerName) (Database, error) {
	uri, err := url.Parse(dataSourceName)
	if err != nil {
		return nil, err
	}
	switch uri.Scheme {
	case "postgres":
		return nil, fmt.Errorf("Cannot use postgres implementation")
	case "mysql":
		return nil, fmt.Errorf("Cannot use mysql implementation")
	case "file":
		return sqlite3.NewPublicRoomsServerDatabase(dataSourceName, localServerName)
	default:
		return nil, fmt.Errorf("Cannot use postgres implementation")
	}
}
