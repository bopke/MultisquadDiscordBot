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

type ColoredUser struct {
	Id                 int       `db:"id,primarykey,autoincrement"`
	DiscordID          string    `db:"discord_id,size:255"`
	Color              string    `db:"color,size:255"`
	Valid              bool      `db:"valid"`
	RoleId             string    `db:"role_id,size:255"`
	ExpirationDate     time.Time `db:"expiration_date"`
	NotifiedExpiration bool      `db:"notified_expiration"`
}

type Raid struct {
	Id        int          `db:"id,primarykey,autoincrement"`
	IssuerId  string       `db:"issuer_id,size:255"`
	ChannelId string       `db:"channel_id,size:255"`
	MessageId string       `db:"message_id,size:255"`
	StartTime time.Time    `db:"start_time"`
	Duration  int          `db:"duration"`
	EndTime   sql.NullTime `db:"end_time"`
}

type ChannelPermissions struct {
	Id                         int    `db:"id,primarykey,autoincrement"`
	RaidId                     int    `db:"raid_id"`
	ChannelId                  string `db:"channel_id,size:255"`
	EveryonePermissionsDenied  int    `db:"permissions_denied"`
	EveryonePermissionsAllowed int    `db:"permissions_allowed"`
}

type Money struct {
	Id     int    `db:"id,primarykey,autoincrement"`
	UserId string `db:"user_id,size:255"`
	Amount int    `db:"amount"`
}

var DbMap gorp.DbMap

//inicjujemy bazę danych
func InitDB() {
	mysqlConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4,utf8", Config.MysqlLogin, Config.MysqlPassword, Config.MysqlHost, Config.MysqlPort, Config.MysqlDatabase)
	db, err := sql.Open("mysql", mysqlConnectionString)
	if err != nil {
		log.Panic("Błąd połączenia z bazą danych!\n" + err.Error())
	}
	// DbMap to schemat bazy danych, dialect oznacza jaką dokładnie baze chcemy używać - mysql, sqlite, postgresql itp...
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	//dodajemy mapie tabelę której struktura jest zdefiniowana wyżej.
	DbMap.AddTableWithName(LinkedUsers{}, "LinkedUsers").SetKeys(true, "id")
	DbMap.AddTableWithName(ColoredUser{}, "ColoredUsers").SetKeys(true, "id")
	DbMap.AddTableWithName(ChannelPermissions{}, "ChannelsPermissions").SetKeys(true, "id")
	DbMap.AddTableWithName(Raid{}, "Raids").SetKeys(true, "id")
	DbMap.AddTableWithName(Money{}, "Money").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("InitDB DbMap.CreateTablesIfNotExists() " + err.Error())
	}
}
