package config

import (
	"encoding/json"
	"os"
)

var (
	MysqlString           string
	DiscordToken          string
	SteamApiToken         string
	ReactionId            = "<:ohgod:760211746345451581"
	ReportsChannelId      = ""
	VerifiedRolesIds      = []string{}
	MoneyLogChannelId     = "597216492123324442"
	ReactionChannelsId    = []string{"597216492123324442"}
	GuildId               string
	AnnouncementChannelId = "597216492123324442"
	PermittedRolesId      = []string{
		"598140510082826270",
	}
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
