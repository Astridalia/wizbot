package cc

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"c.c/bot/database"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/go-chi/chi"
)

type CC struct {
	Client bot.Client
	Mdb    database.MongoImpl
	Mux    *http.ServeMux
}

func New() *CC {
	return &CC{}
}

func (c *CC) SyncCommands(commands []discord.ApplicationCommandCreate, guildIDs ...snowflake.ID) {
	restClient := c.Client.Rest()

	if len(guildIDs) == 0 {
		if _, err := restClient.SetGlobalCommands(c.Client.ApplicationID(), commands); err != nil {
			log.Fatalf("failed to set global commands: %s", err)
		}
		return
	}

	for _, guildID := range guildIDs {
		if _, err := restClient.SetGuildCommands(c.Client.ApplicationID(), guildID, commands); err != nil {
			log.Fatalf("failed to set guild commands: %s", err)
		}
	}
}

func (c *CC) SetupBot(r handler.Router) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("no discord token provided")
	}

	var err error
	c.Client, err = disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuildMessages|gateway.IntentDirectMessages|gateway.IntentGuildMessageTyping|gateway.IntentDirectMessageTyping|gateway.IntentMessageContent),
			gateway.WithCompress(true),
			gateway.WithPresenceOpts(
				gateway.WithPlayingActivity("loading..."),
				gateway.WithOnlineStatus(discord.OnlineStatusDND),
			),
		),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListeners(r),
		bot.WithEventListenerFunc(c.OnReady),
	)

	if err != nil {
		log.Fatalf("Failed to start bot: %s", err)
	}

	if c.Mdb, err = database.Setup("cc"); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
}

func (c *CC) SetupRoutes(router chi.Router) {
	c.Mux = http.NewServeMux()
	c.Mux.Handle("/", router)
}

func (c *CC) StartAndBlock() {
	if err := c.Client.OpenGateway(context.Background()); err != nil {
		log.Fatalf("Failed to connect to gateway: %s", err)
	}

	defer func() {
		log.Print("Shutting down...")
		c.Mdb.Disconnect()
	}()

	log.Print("Client is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}

func (c *CC) OnReady(_ *events.Ready) {
	log.Println("Bot is ready")
}
