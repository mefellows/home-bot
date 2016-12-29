package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"log"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mefellows/bot/actions"
	"github.com/mefellows/home/models"
	"github.com/nlopes/slack"
)

type User struct {
	Info   slack.User
	Rating int
}

type Token struct {
	Token string `json:"token"`
}

type Message struct {
	ChannelId string
	Timestamp string
	Payload   string
	Rating    int
	User      User
}

type BotCentral struct {
	Channel *slack.Channel
	Event   *slack.MessageEvent
	UserId  string
}

type AttachmentChannel struct {
	Channel      *slack.Channel
	Attachment   *slack.Attachment
	DisplayTitle string
}

type Messages []Message

func (u Messages) Len() int {
	return len(u)
}
func (u Messages) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u Messages) Less(i, j int) bool {
	return u[i].Rating > u[j].Rating
}

const TOP = 20

var (
	api               *slack.Client
	botKey            Token
	botId             string
	botCommandChannel chan *BotCentral
	botReplyChannel   chan AttachmentChannel
)

func handleBotCommands(c chan AttachmentChannel) {
	commands := map[string]string{
		"list": "See the current shopping list",
		"help": "See the available bot commands.",
	}

	var attachmentChannel AttachmentChannel

	for {
		botChannel := <-botCommandChannel
		attachmentChannel.Channel = botChannel.Channel
		commandArray := strings.Fields(botChannel.Event.Text)
		log.Println("[DEBUG] received command", commandArray)
		switch commandArray[1] {
		case "help":
			attachmentChannel.DisplayTitle = "Fortune & Karam with Luck"
			fields := make([]slack.AttachmentField, 0)
			for k, v := range commands {
				fields = append(fields, slack.AttachmentField{
					Title: "<bot> " + k,
					Value: v,
				})
			}
			attachment := &slack.Attachment{
				Pretext: "Bot Command List",
				Color:   "#B733FF",
				Fields:  fields,
			}
			attachmentChannel.Attachment = attachment
			c <- attachmentChannel

		case "list":
			fmt.Println("[INFO] list shopping items")
			listAction := &actions.ShoppingAction{}
			attachment := listAction.RetrieveLatestShoppingList()
			attachmentChannel.Attachment = attachment
			c <- attachmentChannel
		case "complete":
			fmt.Println("[INFO] complete shopping list")
			listAction := &actions.ShoppingAction{}
			attachment := listAction.CompleteList()
			attachmentChannel.Attachment = attachment
			c <- attachmentChannel
		case "add":
			fmt.Println("[INFO] append to shopping list")
			listAction := &actions.ShoppingAction{}
			qty, err := strconv.Atoi(commandArray[2])

			if len(commandArray) < 4 || err != nil {
				attachmentChannel.Attachment = &slack.Attachment{
					Pretext: fmt.Sprintf("Invalid command", err),
					Color:   "#B733FF",
				}

			} else {
				item := models.Item{
					Quantity: qty,
					Name:     strings.Join(commandArray[3:], " "),
				}
				attachment := listAction.AppendToShoppingList(item)
				attachmentChannel.Attachment = attachment
			}

			c <- attachmentChannel
		}
	}
}

func handleBotReply() {
	for {
		ac := <-botReplyChannel
		params := slack.PostMessageParameters{}
		params.AsUser = true
		params.Attachments = []slack.Attachment{*ac.Attachment}
		_, _, errPostMessage := api.PostMessage(ac.Channel.Name, ac.DisplayTitle, params)
		if errPostMessage != nil {
			log.Fatal(errPostMessage)
		}
	}
}

func main() {

	api = slack.New(os.Getenv("SLACK_API_KEY"))
	rtm := api.NewRTM()

	botCommandChannel = make(chan *BotCentral)
	botReplyChannel = make(chan AttachmentChannel)

	go rtm.ManageConnection()
	go handleBotCommands(botReplyChannel)
	go handleBotReply()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				log.Println("connected")
				botId = ev.Info.User.ID
			case *slack.TeamJoinEvent:
				log.Println("Team join event")
			case *slack.MessageEvent:
				log.Println("message")
				channelInfo, err := api.GetChannelInfo(ev.Channel)
				if err != nil {
					log.Fatalln(err)
				}

				botCentral := &BotCentral{
					Channel: channelInfo,
					Event:   ev,
					UserId:  ev.User,
				}

				if ev.Type == "message" && strings.HasPrefix(ev.Text, "<@"+botId+">") {
					botCommandChannel <- botCentral
				}

				log.Println("reaction")

			case *slack.ReactionRemovedEvent:
				log.Println("reaction removed")

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				// Ignore other events..
				//fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}
