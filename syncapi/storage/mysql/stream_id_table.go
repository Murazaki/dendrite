package mysql

import (
	"context"
	"database/sql"

	"github.com/matrix-org/dendrite/internal"
	"github.com/matrix-org/dendrite/syncapi/types"
)

const streamIDTableSchema = `
-- Global stream ID counter, used by other tables.
CREATE TABLE IF NOT EXISTS syncapi_stream_id (
  stream_name TEXT NOT NULL PRIMARY KEY,
  stream_id BIGINT DEFAULT 0,

  UNIQUE(stream_name)
);
INSERT INTO syncapi_stream_id (stream_name, stream_id) VALUES ("global", 0)
  ON CONFLICT DO NOTHING;
`

const increaseStreamIDStmt = "" +
	"UPDATE syncapi_stream_id SET stream_id = stream_id + 1 WHERE stream_name = $1"

const selectStreamIDStmt = "" +
	"SELECT stream_id FROM syncapi_stream_id WHERE stream_name = $1"

type streamIDStatements struct {
	increaseStreamIDStmt *sql.Stmt
	selectStreamIDStmt   *sql.Stmt
}

func (s *streamIDStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(streamIDTableSchema)
	if err != nil {
		return
	}
	if s.increaseStreamIDStmt, err = db.Prepare(increaseStreamIDStmt); err != nil {
		return
	}
	if s.selectStreamIDStmt, err = db.Prepare(selectStreamIDStmt); err != nil {
		return
	}
	return
}

func (s *streamIDStatements) nextStreamID(ctx context.Context, txn *sql.Tx) (pos types.StreamPosition, err error) {
	increaseStmt := internal.TxStmt(txn, s.increaseStreamIDStmt)
	selectStmt := internal.TxStmt(txn, s.selectStreamIDStmt)
	if _, err = increaseStmt.ExecContext(ctx, "global"); err != nil {
		return
	}
	if err = selectStmt.QueryRowContext(ctx, "global").Scan(&pos); err != nil {
		return
	}
	return
}
