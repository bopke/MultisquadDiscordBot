package main

import (
	"encoding/json"
	"log"
	"os"
)

// struktura konfiguracji
type Configuration struct {
	MysqlLogin                    string   `json:"mysql_login"`
	MysqlPassword                 string   `json:"mysql_password"`
	MysqlDatabase                 string   `json:"mysql_database"`
	MysqlHost                     string   `json:"mysql_host"`
	MysqlPort                     int      `json:"mysql_port"`
	SteamCommandName              string   `json:"steam_command_name"`
	StatusCommandName             string   `json:"status_command_name"`
	MinecraftCommandName          string   `json:"minecraft_command_name"`
	VipCommandName                string   `json:"vip_command_name"`
	VipDefaultLength              int      `json:"default_vip_length"`
	ColorCommandName              string   `json:"color_command_name"`
	ColorDefaultLength            int      `json:"default_color_length"`
	ColorRoleHierarchyLimiterRole string   `json:"color_role_hierarchy_limiter_role"`
	DiscordToken                  string   `json:"discord_token"`
	SteamApiToken                 string   `json:"steam_api_token"`
	PermittedRoleName             string   `json:"permitted_role_name"`
	AdminRoleName                 string   `json:"admin_role_name"`
	ServerId                      string   `json:"server_id"`
	AnnouncementChannelId         string   `json:"announcement_channel_id"`
	AllowedNicknameChars          string   `json:"allowed_nickname_chars"`
	ChangeBotNicknames            bool     `json:"change_bot_nicknames"`
	RulesMessageId                string   `json:"rules_message_id"`
	RulesAgreementEmojiName       string   `json:"rules_agreement_emoji_name"`
	VerifiedRolesIds              []string `json:"verified_roles_ids"`
}

// zmienna globalna która będzie przechowywać konfigurację
var Config = new(Configuration)

// "metoda" struktury konfiguracji umożliwiająca jej załadowanie
func (c *Configuration) load() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Panic(err)
	}
	defer configFile.Close()
	// dekodujemy całą zawartość pliku na utworzoną strukturę
	err = json.NewDecoder(configFile).Decode(c)
	if err != nil {
		log.Panic("loadConfig Decoder.Decode(c) " + err.Error())
	}
	return
}
