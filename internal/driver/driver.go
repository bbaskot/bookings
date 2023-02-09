package driver

import (
	"database/sql"
	"time"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB struct {
	SQL *sql.DB
}

var DbCon = &DB{}

const maxOpenDbConn=10
const maxIdleDbConn=5
const maxDbLifetime=5*time.Minute


func ConnectSql(dsn string) (*DB,error){
	d, err:=newDatabase(dsn)
	if err!=nil{
		panic(err)
	}
	d.SetConnMaxLifetime(maxDbLifetime)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetMaxOpenConns(maxOpenDbConn)
	DbCon.SQL=d
	if err=DbCon.SQL.Ping();err!=nil{
		return nil,err
	}
	return DbCon,nil

}

func newDatabase(dsn string)(*sql.DB,error){
	db,err:=sql.Open("pgx",dsn)
	if err!=nil{
		return nil,err
	}
	if err=db.Ping();err!=nil{
		return nil,err
	}
	return db,nil
}