package main

import (
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	println("Starting")
	db, err := CreateBoardsDB()
	if err != nil {
		println(err.Error())
		return
	}

	defer func(db *BoardsDB) {
		err := db.Close()
		if err != nil {
			println(err.Error())
		}
	}(db)

	err = db.SetDeletedAll()
	if err != nil {
		println(err.Error())
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
			if existingBoard != nil {
				fmt.Printf("Update: %+v\n", newBoard)
				err = db.Update(newBoard)
				if err != nil {
					println(err.Error())
				}
			} else {
				fmt.Printf("Insert: %+v\n", newBoard)
				_, err = db.Insert(newBoard)
				if err != nil {
					println(err.Error())
				}
			}
		})

		c.Visit(e.ChildAttr("a.next.page-numbers", "href"))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	err = c.Visit("https://www.slosurf.com/ad-category/surf/deske-2/")
	if err != nil {
		fmt.Println(err)
		return
	}

}
