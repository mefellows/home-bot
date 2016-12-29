package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/mefellows/home/models"
	"github.com/nlopes/slack"
)

type ShoppingAction struct{}

var hostname string

func init() {
	hostname = os.Getenv("HOME_API_HOSTNAME")
}

// CompleteList appends an item to the latest list
func (l *ShoppingAction) CompleteList() *slack.Attachment {
	var fields []slack.AttachmentField

	client := &http.Client{}
	r, _ := http.NewRequest("PUT", fmt.Sprintf("%s/shopping/list/complete", hostname), nil)
	r.Header.Set("Content-Type", "application/json")
	res, err := client.Do(r)

	if err != nil || res.StatusCode > 200 {
		fields = []slack.AttachmentField{
			slack.AttachmentField{
				Title: fmt.Sprintf("Unable to complete the list (%d)", res.StatusCode),
				Value: "",
				Short: true,
			},
		}
		return &slack.Attachment{
			Pretext: "Shopping List Not updated!",
			Color:   "#0a84c1",
			Fields:  fields,
		}
	}

	fields = []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Item added",
			Value: "",
			Short: true,
		},
	}
	return &slack.Attachment{
		Pretext: "Shopping List Updated!",
		Color:   "#0a84c1",
		Fields:  fields,
	}
}

// AppendToShoppingList appends an item to the latest list
func (l *ShoppingAction) AppendToShoppingList(item models.Item) *slack.Attachment {
	var fields []slack.AttachmentField

	data, err := json.Marshal(&item)
	if err != nil {
		fields = []slack.AttachmentField{
			slack.AttachmentField{
				Title: "Unable to add item",
				Value: err.Error(),
				Short: true,
			},
		}
		return &slack.Attachment{
			Pretext: "Shopping List Not updated!",
			Color:   "#0a84c1",
			Fields:  fields,
		}

	}
	log.Println("[DEBUG] appending item", item)
	log.Println("[DEBUG] appending item", string(data))
	client := &http.Client{}
	r, _ := http.NewRequest("POST", fmt.Sprintf("%s/shopping/list/append", hostname), bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	res, err := client.Do(r)

	if err != nil || res.StatusCode > 200 {
		fields = []slack.AttachmentField{
			slack.AttachmentField{
				Title: fmt.Sprintf("Item not added added (%d)", res.StatusCode),
				Value: "",
				Short: true,
			},
		}
		return &slack.Attachment{
			Pretext: "Shopping List Not updated!",
			Color:   "#0a84c1",
			Fields:  fields,
		}
	}

	fields = []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Item added",
			Value: "",
			Short: true,
		},
	}
	return &slack.Attachment{
		Pretext: "Shopping List Updated!",
		Color:   "#0a84c1",
		Fields:  fields,
	}
}

// RetrieveLatestShoppingList gets the latest shopping list
func (l *ShoppingAction) RetrieveLatestShoppingList() *slack.Attachment {

	resp, err := http.Get(fmt.Sprintf("%s/shopping/list", hostname))

	if err == nil {
		defer resp.Body.Close()
		var list models.List
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("[INFO] retrieved shopping list")
		if err := json.Unmarshal(body, &list); err != nil {
			log.Println("[ERROR]", err.Error())
			return nil
		}

		fields := make([]slack.AttachmentField, len(list.Items))
		for i, item := range list.Items {
			fields[i] = slack.AttachmentField{
				Title: item.Name,
				Value: fmt.Sprintf("%d x %s %s", item.Quantity, item.Name, item.Description),
				Short: false,
			}
		}

		attachment := &slack.Attachment{
			Pretext: fmt.Sprintf("Shopping List (updated %s)", list.UpdatedAt),
			Color:   "#0a84c1",
			Fields:  fields,
		}
		return attachment
	}
	return nil
}
