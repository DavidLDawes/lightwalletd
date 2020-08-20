// Package common Copyright (c) 2019-2020 The Zcash developers
// Copyright 2020 The VerusCoin Developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or https://www.opensource.org/licenses/mit-license.php .
package common

import (
	"context"
	"fmt"

	"github.com/asherda/lightwalletd/walletrpc"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zcash/lightwalletd/common"
)

// DbConfig is used by root.go to ser the DB info
type DbConfig struct {
	SQLHost string
	SQLPort uint
	SQLUser string
	SQLPW   string
}

// GetDBPool gets a pgx Pool ready for the ingestor while checking for errors
func GetDBPool(cfg DbConfig) *pgxpool.Pool {
	// TODO: switch Postgres data to command line options
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		cfg.SQLHost, cfg.SQLPort, cfg.SQLUser, cfg.SQLPW)

	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		common.Log.WithFields(logrus.Fields{
			"error":   err,
			"sqlInfo": psqlInfo,
		}).Fatal("unable to parse psqlInfo for DB connection: %s\n\n", psqlInfo)
	}

	dbPool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		common.Log.WithFields(logrus.Fields{
			"error":   err,
			"sqlInfo": psqlInfo,
		}).Fatal("unable to configure connections for the pool using the poolconfig built from %s", psqlInfo)
	}
	return dbPool
}

func persistToDB(db *pgxpool.Conn, height uint64, hash []byte, prevHash []byte, time uint32, header []byte, vtx []*walletrpc.CompactTx) (string, error) {
	// Add it to PostgreSQL DB
	var err error = nil

	// First add the blocks record
	// until we fix the header, put a fake header in place if needed
	// (it's always needed because we have not fixed the header yet)
	var tempHeader []byte = nil
	if header == nil {
		tempHeader = []byte("Missing a header still")
	} else {
		tempHeader = header
	}

	// All or nothing: get all related records updated, or none at all
	blocktx, err := db.Begin(context.Background())
	if err != nil {
		return fmt.Sprint("unable to open a transaction, failed to write block to DB at height %1", height), err
	}

	commandTag, err := blocktx.Exec(context.Background(), "SELECT COUNT(*) FROM blocks WHERE height=$1;", height)
	if commandTag.RowsAffected() == 1 {
		// Exisating record - a reorg, or restart after losing cache but not DB
		// first get rid of the old block and the related stuff (txs, spends, outputs)
		commandTag, err = db.Exec(context.Background(), "DELETE FROM blocks WHERE height = $1;", height)
		if err != nil {
			blocktx.Rollback(context.Background())
			return fmt.Sprintf("Failed to cascade delete existing block, unable to save block in DB at height ", height), err
		}
	}

	// Either this block did not exist already in the DB, or (if it did)
	// we cascade deleted it so it no longer exists now, anyway
	commandTag, err = blocktx.Exec(context.Background(), "INSERT INTO blocks(height, hash, prev_hash, time, header) VALUES ($1, $2, $3, $4, $5);", height, hash, prevHash, time, tempHeader)
	if err != nil {
		// That failed, so end the DB transaction
		blocktx.Rollback(context.Background())

		// Block already exists, reorg possible so replace it
		// Start a DB transaction again, this time deleting first
		blocktx, err := db.Begin(context.Background())
		if err != nil {
			return fmt.Sprint("unable to open a transaction after insert failed, failed to write block to DB at height %1", height), err
		}
		// first get rid of it and related stuff
		commandTag, err = db.Exec(context.Background(), "DELETE FROM blocks WHERE height = $1;", height)

		if err != nil {
			blocktx.Rollback(context.Background())
			return fmt.Sprintf("Failed to add new block, failed to delete existing block, unable to save block in DB at height %1", height), err
		}

		// deleted OK, (cascading through tx and it's outputs and spends, so put
		// it back in now
		_, err = blocktx.Exec(context.Background(), "INSERT INTO blocks(height, hash, prev_hash, time, header) VALUES ($1, $2, $3, $4, $5);", height, hash, prevHash, time, tempHeader)
		if err != nil {
			blocktx.Rollback(context.Background())
			return fmt.Sprintf("Unable to insert record into DB blocks after deleting at height %1", height), err
		}
	} else {
		result, err := checkExecResult(err, commandTag, "block", height)
		if err != nil {
			blocktx.Rollback(context.Background())
			return result, err
		}
	}

	// Now handle the TX array - put it in it's own table with a reference
	// to the height of the related block
	for _, tx := range vtx {
		commandTag, err := blocktx.Exec(context.Background(), "INSERT INTO tx(index, height, hash, fee) VALUES ($1, $2, $3, $4);", tx.Index, height, tx.GetHash(), tx.GetFee())
		result, err := checkExecResult(err, commandTag, "tx", height)
		if err != nil {
			blocktx.Rollback(context.Background())
			return result, err
		}

		// Within each tx, handle the spend array - put it in it's own table
		// with a reference to the TX hash of the related tx
		for _, spend := range tx.GetSpends() {
			commandTag, err := blocktx.Exec(context.Background(), "INSERT INTO spend(tx_hash, nf) VALUES ($1, $2);", tx.GetHash(), spend.GetNf())
			result, err := checkExecResult(err, commandTag, "spend", height)
			if err != nil {
				blocktx.Rollback(context.Background())
				return result, err
			}
		}

		// Within each tx, handle the output array - put it in it's own table
		// with a reference to the TX hash of the related tx
		for _, output := range tx.GetOutputs() {
			commandTag, err := blocktx.Exec(context.Background(), "INSERT INTO output(tx_hash, cmu, epk, ciphertext) VALUES ($1, $2, $3, $4);", tx.GetHash(), output.GetCmu(), output.GetEpk(), output.GetCiphertext())
			result, err := checkExecResult(err, commandTag, "output", height)
			if err != nil {
				blocktx.Rollback(context.Background())
				return result, err
			}
		}
	}
	err = blocktx.Commit(context.Background())
	if err != nil {
		return fmt.Sprint("Failed to commit block and related table updates at height ", height), err
	}

	return "", nil
}

func checkExecResult(err error, tag pgconn.CommandTag, table string, height uint64) (string, error) {
	if err != nil {
		// Record already exists, error result
		return fmt.Sprintf("Failed to add new %s, failed to delete existing block, unable save block in DB at height %d", table, height), err
	} else {
		if tag.RowsAffected() < 1 {
			return fmt.Sprintf("insert into %s affected 0 rows, should be 1, at height %1", table, height), errors.New("insert into block did not affect any rows")
		} else {
			if tag.RowsAffected() > 1 {
				return fmt.Sprint("insert into %s affected too many rows - %d -  should be 1 - at height %d", table, tag.RowsAffected, height), errors.New("insert into blocks affected too many rows")
			} else {
				return "", nil
			}
		}
	}
}
