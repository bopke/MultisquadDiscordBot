package database

import (
	"database/sql"
	"github.com/bopke/MultisquadDiscordBot/config"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v2"
)

var DbMap gorp.DbMap

func InitMysql() error {
	db, err := sql.Open("mysql", config.MysqlString)
	if err != nil {
		return err
	}
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	DbMap.AddTableWithName(LinkedUsers{}, "LinkedUsers").SetKeys(true, "id")
	DbMap.AddTableWithName(ColoredUser{}, "ColoredUsers").SetKeys(true, "id")
	DbMap.AddTableWithName(ChannelPermissions{}, "ChannelsPermissions").SetKeys(true, "id")
	DbMap.AddTableWithName(Raid{}, "Raids").SetKeys(true, "id")
	DbMap.AddTableWithName(Money{}, "Money").SetKeys(true, "id")
	DbMap.AddTableWithName(ShopLog{}, "ShopLogs").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		return err
	}
	return nil
}
