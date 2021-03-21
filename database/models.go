package database

import (
	"database/sql"
	"time"
)

type LinkedUsers struct {
	Id                 int            `db:"id,primarykey,autoincrement"`
	DiscordId          string         `db:"discord_id,size:255"`
	SteamId64          sql.NullString `db:"steam_id,size:255"`
	Valid              bool           `db:"valid"`
	ExpirationDate     time.Time      `db:"expiration_date"`
	MinecraftNickname  sql.NullString `db:"minecraft_nickname,size:255"`
	NotifiedExpiration bool           `db:"notified_expiration"`
}

type ShopLog struct {
	Id        int       `db:"id,primarykey,autoincrement"`
	DiscordId string    `db:"discord_id,size:255"`
	Item      string    `db:"item,size:255"`
	Price     int       `db:"price"`
	Date      time.Time `db:"date"`
}

type ColoredUser struct {
	Id                 int       `db:"id,primarykey,autoincrement"`
	DiscordId          string    `db:"discord_id,size:255"`
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
