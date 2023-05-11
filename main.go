package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func sendMessageToMessenger(message string) {
	messengerUrlString := os.Getenv("MESSENGER_URL")
	messengerAccessToken := os.Getenv("MESSENGER_ACCESS_TOKEN")
	messengerUserId := os.Getenv("MESSENGER_USER_ID")

	if len(messengerUrlString) == 0 {
		log.Println("Environment variable \"MESSENGER_URL\" is not set!")
		return
	}

	if len(messengerAccessToken) == 0 {
		log.Println("Environment variable \"ACCESS_TOKEN\" is not set!")
		return
	}

	if len(messengerUserId) == 0 {
		log.Println("Environment variable \"USER_ID\" is not set!")
		return
	}

	messengerUrl, err := url.Parse(messengerUrlString)
	if err != nil {
		log.Fatalln(err)
		return
	}

	query := messengerUrl.Query()
	messageJson, _ := json.Marshal(map[string]string{
		"text": message,
	})
	query.Set("message", string(messageJson))
	query.Set("messaging_type", "RESPONSE")
	recipientJson, _ := json.Marshal(map[string]string{
		"id": messengerUserId,
	})
	query.Set("recipient", string(recipientJson))
	query.Set("access_token", messengerAccessToken)

	messengerUrl.RawQuery = query.Encode()
	resp, err := http.Post(
		messengerUrl.String(),
		"application/json; charset=UTF-8",
		nil,
	)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(body))
}

func main() {
	db, err := CreateBoardsDB()
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer func(db *BoardsDB) {
		err := db.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(db)

	err = db.SetDeletedAll()
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.slosurf.com"),
	)

	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {

		e.ForEach("article", func(i int, article *colly.HTMLElement) {
			var newBoard Board

			postId, err := strconv.ParseInt(article.Attr("data-id"), 10, 64)
			if err == nil {
				newBoard.PostId = postId
			} else {
				return
			}

			newBoard.Title = article.ChildAttr("h2.h4.entry-title a", "title")
			if newBoard.Title == "" || strings.Contains(strings.ToLower(newBoard.Title), "kupim") {
				return
			}

			newBoard.Link = article.ChildAttr("h2.h4.entry-title a", "href")

			priceString := strings.ReplaceAll(article.ChildText("div.price-wrap span.tag-head span.post-price"), "€", "")
			priceString = strings.ReplaceAll(priceString, ".", "")
			priceString = strings.ReplaceAll(priceString, ",", ".")
			price, err := strconv.ParseFloat(priceString, 64)
			if err == nil {
				newBoard.Price = price
			}

			selection := article.DOM.Find("div.entry-content.subheader span")
			selection.Each(func(i int, selection *goquery.Selection) {
				if selection.HasClass("cfd_volume") {
					selection = selection.Next()
					liters, err := strconv.ParseFloat(selection.Nodes[0].FirstChild.Data, 64)
					if err == nil {
						newBoard.Liters = liters
					}
				} else if selection.HasClass("cfd_size") {
					selection = selection.Next()
					length, err := strconv.ParseFloat(selection.Nodes[0].FirstChild.Data, 64)
					if err == nil {
						newBoard.Length = length
					}
				} else if selection.HasClass("cfd_weight") {
					selection = selection.Next()
					weight, err := strconv.ParseFloat(selection.Nodes[0].FirstChild.Data, 64)
					if err == nil {
						newBoard.Weight = weight
					}
				}
			})

			newBoard.Description = article.ChildText("div.entry-content.subheader")

			existingBoard, err := db.GetByPostId(newBoard.PostId)
			if err != nil {
				log.Println(err.Error())
			}

			if existingBoard != nil {
				//log.Printf("Update: %+v\n", newBoard)
				err = db.Update(newBoard)
				if err != nil {
					log.Println(err.Error())
				}
			} else {
				log.Printf("Insert: %+v\n", newBoard)
				if newBoard.Price != 0 {
					sendMessageToMessenger(fmt.Sprintf("New board: %s (%.2f €)\n%s", newBoard.Title, newBoard.Price, newBoard.Link))
				} else {
					sendMessageToMessenger(fmt.Sprintf("New board: %s\n%s", newBoard.Title, newBoard.Link))
				}
				_, err = db.Insert(newBoard)
				if err != nil {
					log.Println(err.Error())
				}
			}
		})

		c.Visit(e.ChildAttr("a.next.page-numbers", "href"))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	err = c.Visit("https://www.slosurf.com/ad-category/surf/deske-2/")
	if err != nil {
		log.Println(err)
		return
	}

}
