package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"strings"
)

var (
	dbInstance *sql.DB
)

type Tx struct {
	Id               uint64          `db:"id"`
	BlockNo          int64           `db:"block_no"`
	BlockTime        int64           `db:"block_time"`
	BlockHash        string          `db:"block_hash"`
	TxHash           string          `db:"tx_hash"`
	Amount           decimal.Decimal `db:"amount"`
	AddressFrom      string          `db:"address_from"`
	Destination      string          `db:"destination"`
	Token            string          `db:"token"`
	SignatureChainId string          `db:"signature_chain_id"`
	HyperLiquidChain string          `db:"hyper_liquid_chain"`
	ActionType       string          `db:"type"`
	Error            string          `db:"error"`
}

func init() {
	var err error
	dbInstance, err = sql.Open("sqlite3", "./hype-sync.db")
	if err != nil {
		panic(err)
	}
	err = CreateTableTx()
	if err != nil {
		panic(err)
	}
}

func CreateTableTx() error {
	sqlStr := `
    CREATE TABLE IF NOT EXISTS tx(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        block_no INTEGER not NULL default 0,
        block_time INTEGER not NULL default 0,
        block_hash VARCHAR(64) not NULL default '',
        tx_hash VARCHAR(64) not NULL default '',
        amount decimal(65,18) not null default 0 ,
        address_from VARCHAR(64) not NULL default '',
        destination VARCHAR(64) not NULL default '',
        token VARCHAR(64) not NULL default '',
        signature_chain_id VARCHAR(64) not NULL default '',
        hyper_liquid_chain VARCHAR(64) not NULL default '',
        action_type VARCHAR(64) not NULL default '',
        error VARCHAR(64) not NULL default ''
    );
    `
	_, err := dbInstance.Exec(sqlStr)
	if err != nil {
		return err
	}

	row, err := FindLastTx()
	if err == nil && row == nil {
		sqlStr = `create unique index uni_block_no on tx (block_no);`
		_, err = dbInstance.Exec(sqlStr)
		fmt.Println(err)
		AddTx(&Tx{
			BlockNo:          434565834,
			BlockTime:        1735024473924,
			BlockHash:        "0x6e553dfc15eb03b24a2f48f0613b88d71eab49cb233d22de719fa251b301503d",
			TxHash:           "0xac00cc86004d2d7325b50419e6f2ca018200f85e736f625243e1bea81123620c",
			Amount:           decimal.RequireFromString("0.1"),
			AddressFrom:      "0x633a84ee0ab29d911e5466e5e1cb9cdbf5917e72",
			Destination:      "0x63e67af80a212b832a8dc8a30faaed8b0dde6b6b",
			Token:            "HYPE:0x0d01dc56dcaaca66ad901c959b4011ec",
			SignatureChainId: "0xa4b1",
			HyperLiquidChain: "Mainnet",
			ActionType:       "spotSend",
			Error:            "",
		})
		//sqlStr = `create index txid on tx (tx_hash);`
		//_, err = dbInstance.Exec(sqlStr)
		//fmt.Println(err)
	}
	return nil
}

func FindLastTx() (*Tx, error) {
	var row Tx
	err := dbInstance.QueryRow("select * from `tx` order by `id` desc limit 1").Scan(
		&row.Id,
		&row.BlockNo,
		&row.BlockTime,
		&row.BlockHash,
		&row.TxHash,
		&row.Amount,
		&row.AddressFrom,
		&row.Destination,
		&row.Token,
		&row.SignatureChainId,
		&row.HyperLiquidChain,
		&row.ActionType,
		&row.Error,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func FindByBlockNum(num int64) ([]*Tx, error) {
	var ret []*Tx
	rows, err := dbInstance.Query("select * from `tx` where block_no=?", num)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var row Tx
		err = rows.Scan(
			&row.Id,
			&row.BlockNo,
			&row.BlockTime,
			&row.BlockHash,
			&row.TxHash,
			&row.Amount,
			&row.AddressFrom,
			&row.Destination,
			&row.Token,
			&row.SignatureChainId,
			&row.HyperLiquidChain,
			&row.ActionType,
			&row.Error,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		ret = append(ret, &row)
	}
	return ret, nil
}

func AddTx(row *Tx) error {
	stmt, err := dbInstance.Prepare("INSERT INTO `tx`(`block_no`, `block_time`, `block_hash`, `tx_hash`, `amount`, `address_from`, `destination`, `token`, `signature_chain_id`, `hyper_liquid_chain`, `action_type`, `error`) values(?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(
		row.BlockNo,
		row.BlockTime,
		row.BlockHash,
		row.TxHash,
		row.Amount,
		row.AddressFrom,
		row.Destination,
		row.Token,
		row.SignatureChainId,
		row.HyperLiquidChain,
		row.ActionType,
		row.Error,
	)
	if err != nil {
		if strings.EqualFold("UNIQUE constraint failed: main_chain.block_no", err.Error()) {
			return nil
		}
		return err
	}

	id, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if id == 0 {
		return errors.New("insert 0 row")
	}

	return nil
}
