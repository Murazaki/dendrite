// Copyright 2017-2018 New Vector Ltd
// Copyright 2019-2020 The Matrix.org Foundation C.I.C.
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

package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/matrix-org/dendrite/roomserver/types"
)

const stateSnapshotSchema = `
  CREATE TABLE IF NOT EXISTS roomserver_state_snapshots (
    state_snapshot_nid INTEGER PRIMARY KEY AUTOINCREMENT,
    room_nid INTEGER NOT NULL,
    state_block_nids TEXT NOT NULL
  );
`

const insertStateSQL = `
	INSERT INTO roomserver_state_snapshots (room_nid, state_block_nids)
	  VALUES ($1, $2);
	SELECT state_snapshot_nid FROM roomserver_state_snapshots
		WHERE rowid = last_insert_rowid();
`

// Bulk state data NID lookup.
// Sorting by state_snapshot_nid means we can use binary search over the result
// to lookup the state data NIDs for a state snapshot NID.
const bulkSelectStateBlockNIDsSQL = "" +
	"SELECT state_snapshot_nid, state_block_nids FROM roomserver_state_snapshots" +
	" WHERE state_snapshot_nid IN ($1) ORDER BY state_snapshot_nid ASC"

type stateSnapshotStatements struct {
	insertStateStmt              *sql.Stmt
	bulkSelectStateBlockNIDsStmt *sql.Stmt
}

func (s *stateSnapshotStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(stateSnapshotSchema)
	if err != nil {
		return
	}

	return statementList{
		{&s.insertStateStmt, insertStateSQL},
		{&s.bulkSelectStateBlockNIDsStmt, bulkSelectStateBlockNIDsSQL},
	}.prepare(db)
}

func (s *stateSnapshotStatements) insertState(
	ctx context.Context, roomNID types.RoomNID, stateBlockNIDs []types.StateBlockNID,
) (stateNID types.StateSnapshotNID, err error) {
	nids := make([]int64, len(stateBlockNIDs))
	for i := range stateBlockNIDs {
		nids[i] = int64(stateBlockNIDs[i])
	}
	err = s.insertStateStmt.QueryRowContext(ctx, int64(roomNID), pq.Int64Array(nids)).Scan(&stateNID)
	return
}

func (s *stateSnapshotStatements) bulkSelectStateBlockNIDs(
	ctx context.Context, stateNIDs []types.StateSnapshotNID,
) ([]types.StateBlockNIDList, error) {
	nids := make([]int64, len(stateNIDs))
	for i := range stateNIDs {
		nids[i] = int64(stateNIDs[i])
	}
	rows, err := s.bulkSelectStateBlockNIDsStmt.QueryContext(ctx, pq.Int64Array(nids))
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint: errcheck
	results := make([]types.StateBlockNIDList, len(stateNIDs))
	i := 0
	for ; rows.Next(); i++ {
		result := &results[i]
		var stateBlockNIDs pq.Int64Array
		if err := rows.Scan(&result.StateSnapshotNID, &stateBlockNIDs); err != nil {
			return nil, err
		}
		result.StateBlockNIDs = make([]types.StateBlockNID, len(stateBlockNIDs))
		for k := range stateBlockNIDs {
			result.StateBlockNIDs[k] = types.StateBlockNID(stateBlockNIDs[k])
		}
	}
	if i != len(stateNIDs) {
		return nil, fmt.Errorf("storage: state NIDs missing from the database (%d != %d)", i, len(stateNIDs))
	}
	return results, nil
}