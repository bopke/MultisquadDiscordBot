package config

import (
	"encoding/json"
	"os"
)

var (
	MysqlString      string
	DiscordToken     string
	SteamApiToken    string
	ReactionId       = "<:lov:582679593060663315"
	ReportsChannelId = "581927703406444555"
	VerifiedRolesIds = []string{"610961053949624320",
		"611162396215738380",
		"610961378966110248",
		"582166692831166485",
		"610962640164093952",
		"662670498949365771"}
	MoneyLogChannelId  = "613027719902527501"
	ReactionChannelsId = []string{"664518622445568060",
		"588855162077184011"}
	GuildId               string
	AnnouncementChannelId = "580139853933707265"
	PermittedRolesId      = []string{}
)

func Load() error {
	type Configuration struct {
		MysqlString   string `json:"mysql_string"`
		DiscordToken  string `json:"discord_token"`
		SteamApiToken string `json:"steam_api_token"`
		GuildId       string `json:"guild_id"`
	}
	var c Configuration
	configFile, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer configFile.Close()
	// dekodujemy całą zawartość pliku na utworzoną strukturę
	err = json.NewDecoder(configFile).Decode(&c)
	if err != nil {
		return err
	}
	MysqlString = c.MysqlString
	DiscordToken = c.DiscordToken
	SteamApiToken = c.SteamApiToken
	GuildId = c.GuildId
	return nil
}
