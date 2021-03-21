package main

import (
	"github.com/bopke/MultisquadDiscordBot/colors"
	"github.com/bopke/MultisquadDiscordBot/commands"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/nicks"
	"github.com/bopke/MultisquadDiscordBot/vip"
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// sesja połączenia z botem musi być globalna, żeby cykliczne zadanie sprawdzania użytkowników mogło działać
var session *discordgo.Session

//stan dodawania steamId do bazy, używane przy funkcji łączenia użytkownika discorda z kontem steam.
type State byte

const (
	INSERTED State = iota
	UPDATED
	ERROR
)

func main() {
	log.Println("Warming up...")
	err := config.Load()
	if err != nil {
		log.Panicln("Unable to load config", err)
		return
	}
	Locale.load()
	err = database.InitMysql()
	if err != nil {
		log.Panicln("Unable to connect to database", err)
		return
	}
	session, err = discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		panic(err)
	}

	commands.Init()
	//	session.AddHandler(commands.Listener)
	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnGuildMemberUpdate)
	session.AddHandler(OnMessageReactionAdd)
	session.AddHandler(OnGuildMemberAdd)
	session.AddHandler(OnDMMessageReactionAdd)

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	err = session.Open()
	if err != nil {
		panic(err)
	}

	// cron - narzędzie do cyklicznego wykonywania zadania. Co minutę będzie odpalać funkcję checkUsers.
	c := cron.New()
	_ = c.AddFunc("0 0 * * * *", func() {
		ctx := &context.Context{
			Session: session,
		}
		err := colors.CheckUserColors(ctx)
		if err != nil {
			log.Println("Error while checking users colors", err)
		}
		err = vip.CheckVips(ctx)
		if err != nil {
			log.Println("Error while checking users colors", err)
		}
	})
	_ = c.AddFunc("0 0 0 * * *", func() { rankMoneyAdd("579717933736132620", 300, "") })
	_ = c.AddFunc("0 0 6 * * *", func() { rankMoneyAdd("579717933736132620", 300, "") })
	_ = c.AddFunc("0 0 12 * * *", func() { rankMoneyAdd("579717933736132620", 300, "") })
	_ = c.AddFunc("0 0 13 * * *", func() { rankMoneyAdd("611201074275155969", 10, "") })
	_ = c.AddFunc("0 0 14 * * *", func() { rankMoneyAdd("611202192438853642", 25, "") })
	_ = c.AddFunc("0 0 15 * * *", func() { rankMoneyAdd("651919586190557190", 45, "") })
	_ = c.AddFunc("0 0 16 * * *", func() { rankMoneyAdd("717495454409162812", 100, "") })
	_ = c.AddFunc("0 0 17 * * *", func() { rankMoneyAdd("581900782828257280", 30, "") })
	_ = c.AddFunc("0 0 19 * * *", func() { rankMoneyAdd("586314300927508521", 60, "") })
	_ = c.AddFunc("0 0 18 * * *", func() { rankMoneyAdd("579717933736132620", 300, "") })
	_ = c.AddFunc("0 0 20 * * *", func() { rankMoneyAdd("658394215470071819", 120, "") })
	_ = c.AddFunc("0 0 21 * * *", func() { rankMoneyAdd("719226922432725072", 240, "") })

	//	_ = c.AddFunc("0 39 4 * * *", func() { log.Println("witam shalom");rankMoneyAdd("597569142580576257", 100000, "") })
	c.Start()

	go inits()
	log.Println("Started.")
	// ten kanał powoduje utrzymanie działania programu dopóki nie przyjdzie do niego sygnał od systemu operacyjnego, że pora się zwijać
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	err = session.Close()
	if err != nil {
		panic(err)
	}
}

func inits() {
	_ = vip.CheckVips(&context.Context{Session: session})
	_ = colors.CheckUserColors(&context.Context{Session: session})
	_ = nicks.CheckNicknames(session)
}

//funkcja odpalana cyklicznie, sprawdza czy wszyscy w bazie nadal są na serwerze i czy są na nim vipami.
func checkUsers() {
	_ = vip.CheckVips(&context.Context{Session: session})
	_ = colors.CheckUserColors(&context.Context{Session: session})
	_ = nicks.CheckNicknames(session)
}
