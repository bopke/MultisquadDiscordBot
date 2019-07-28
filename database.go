package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v2"
	"log"
	"time"
)

// struktura tabeli w bazie danych
type LinkedUsers struct {
	Id                 int            `db:"id,primarykey,autoincrement"`
	DiscordID          string         `db:"discord_id,size:255"`
	SteamID64          sql.NullString `db:"steam_id,size:255"`
	Valid              bool           `db:"valid"`
	ExpirationDate     time.Time      `db:"expiration_date"`
	MinecraftNickname  sql.NullString `db:"minecraft_nickname,size:255"`
	NotifiedExpiration bool           `db:"notified_expiration"`
}

var DbMap gorp.DbMap

//inicjujemy bazę danych
func InitDB() {
	mysqlConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", Config.MysqlLogin, Config.MysqlPassword, Config.MysqlHost, Config.MysqlPort, Config.MysqlDatabase)
	db, err := sql.Open("mysql", mysqlConnectionString)
	if err != nil {
		log.Panic("Błąd połączenia z bazą danych!\n" + err.Error())
	}
	// DbMap to schemat bazy danych, dialect oznacza jaką dokładnie baze chcemy używać - mysql, sqlite, postgresql itp...
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	//dodajemy mapie tabelę której struktura jest zdefiniowana wyżej.
	DbMap.AddTableWithName(LinkedUsers{}, "LinkedUsers").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("InitDB DbMap.CreateTablesIfNotExists() " + err.Error())
	}
}
