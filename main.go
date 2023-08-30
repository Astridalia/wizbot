package main

import (
	"flag"
	"log"

	"c.c/bot/cc"
	"c.c/bot/commands"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/go-chi/chi"
)

var (
	shouldSyncCommands *bool
	devMode            *bool
)

func init() {
	shouldSyncCommands = flag.Bool("sync-commands", true, "sync commands to discord")
	devMode = flag.Bool("dev-mode", true, "run in dev mode")
	flag.Parse()
}

func main() {
	c := cc.New()
	r := chi.NewRouter()
	c.SetupRoutes(r)
	cr := handler.New()
	registerCommandHandlers(cr, c)
	c.SetupBot(cr)
	if *shouldSyncCommands {
		syncCommands(c)
	}
	c.StartAndBlock()
}

func registerCommandHandlers(cr handler.Router, c *cc.CC) {
	cr.Command("/ping", commands.HandlePing)
	cr.Route("/wiki", func(r handler.Router) {
		r.Autocomplete("/", commands.HandleFinderAutoComplete(c))
		r.Command("/", commands.HandleFinder(c))
	})
}

func syncCommands(c *cc.CC) {
	var guildIDs []snowflake.ID
	if *devMode {
		guildID := ""
		if guildID == "" {
			log.Fatal("no discord guild id provided")
		}
		guildIDs = append(guildIDs, snowflake.MustParse(guildID))
	}
	c.SyncCommands(commands.Commands, guildIDs...)
}
