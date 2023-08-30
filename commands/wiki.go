package commands

import (
	"context"
	"log"
	"time"

	"c.c/bot/cc"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// finderCommand represents the slash command "/finder" used for searching in the Wizard101 Finder.
var finderCommand = discord.SlashCommandCreate{
	Name:        "wiki",
	Description: "Search the wiki",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:         "query",
			Description:  "The name of the card to search for",
			Required:     true,
			Autocomplete: true,
		},
	},
}

// Card represents the structure of a card object.
type Card struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Image       string             `bson:"image_link"`
	Wiki        string             `bson:"wiki_link"`
	Pack        string             `bson:"pack,omitempty"`
	School      string             `bson:"school,omitempty"`
}

// objectIDFromHex converts a hexadecimal string to a primitive.ObjectID.
func objectIDFromHex(hex string) primitive.ObjectID {
	objectID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		log.Fatal(err)
	}
	return objectID
}

// handleEmbed creates and returns a Discord embed builder for handling errors.
func handleEmbed(err error) discord.EmbedBuilder {
	eb := discord.NewEmbedBuilder()
	if err != nil {
		eb.SetTitle("Error")
		eb.SetDescription(err.Error())
	}
	return *eb
}

// FindOneByID retrieves a card by its ID from the database.
func FindOneByID(c *cc.CC, id string) (Card, error) {
	findOne := c.Mdb.FindOne("tcs", bson.M{"_id": objectIDFromHex(id)})
	var card Card
	err := findOne.Decode(&card)
	return card, err
}

// HandleFinder is the command handler for the "/finder" command.
func HandleFinder(c *cc.CC) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		data := e.SlashCommandInteractionData().String("query")
		card, err := FindOneByID(c, data)
		builder := handleEmbed(err)
		builder.SetAuthor(card.Name, card.Wiki, card.Image)
		builder.SetDescription(card.Description)
		if card.Pack != "" {
			builder.AddField("Pack", card.Pack, true)
		}
		return handleMessage(e, builder)
	}
}

// handleMessage sends the response message containing the card information.
func handleMessage(e *handler.CommandEvent, builder discord.EmbedBuilder) error {
	return e.Respond(discord.InteractionResponseTypeCreateMessage,
		messageBuilder(builder),
	)
}

// messageBuilder constructs a Discord message containing the embed.
func messageBuilder(builder discord.EmbedBuilder) discord.MessageCreate {
	return discord.NewMessageCreateBuilder().
		AddEmbeds(builder.Build()).
		SetEphemeral(true).
		Build()
}

// HandleFinderAutoComplete is the autocomplete handler for the "/finder" command.
func HandleFinderAutoComplete(c *cc.CC) handler.AutocompleteHandler {
	return func(e *handler.AutocompleteEvent) error {
		query := e.AutocompleteInteraction.Data.String("query")

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cursor, err := c.Mdb.Cursor("tcs", query)
		if err != nil {
			log.Printf("Error getting cards: %s", err)
			return handleError(err, e)
		}

		defer cursor.Close(ctx)

		choices := make([]discord.AutocompleteChoice, 0, 25)
		seenNames := make(map[string]struct{})

		for cursor.Next(ctx) {
			var card Card
			err := cursor.Decode(&card)
			if err != nil {
				log.Printf("Error decoding card: %s", err)
				continue
			}

			if _, seen := seenNames[card.Name]; seen || len(choices) >= 25 {
				continue
			}

			seenNames[card.Name] = struct{}{}
			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  card.Name,
				Value: card.ID.Hex(),
			})
		}

		if err := cursor.Err(); err != nil {
			log.Printf("Error during cursor iteration: %s", err)
		}

		return e.Result(choices)
	}
}

// handleError handles errors during autocomplete and returns an empty result.
func handleError(err error, e *handler.AutocompleteEvent) error {
	if err != nil {
		log.Printf("Error getting cards: %s", err)
	}
	return e.Result([]discord.AutocompleteChoice{})
}
