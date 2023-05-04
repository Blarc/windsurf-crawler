package main

import (
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"net/http"
	"strconv"
	"strings"
)

type board struct {
	PostId      int64
	Title       string
	Price       float64
	Liters      float64
	Weight      float64
	Length      float64
	Description string
	Link        string
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("www.slosurf.com"),
	)

	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	// On every an element which has href attribute call callback
	c.OnHTML("body", func(e *colly.HTMLElement) {

		e.ForEach("article", func(i int, article *colly.HTMLElement) {
			var board board

			postId, err := strconv.ParseInt(article.Attr("data-id"), 10, 64)
			if err == nil {
				board.PostId = postId
			} else {
				return
			}

			board.Title = article.ChildAttr("h2.h4.entry-title a", "title")
			if board.Title == "" || strings.Contains(strings.ToLower(board.Title), "kupim") {
				return
			}

			board.Link = article.ChildAttr("h2.h4.entry-title a", "href")

			priceString := strings.ReplaceAll(article.ChildText("div.price-wrap span.tag-head span.post-price"), "€", "")
			priceString = strings.ReplaceAll(priceString, ",", ".")
			price, err := strconv.ParseFloat(priceString, 64)
			if err == nil {
				board.Price = price
			}

			selection := article.DOM.Find("div.entry-content.subheader span")
			selection.Each(func(i int, selection *goquery.Selection) {
				if selection.HasClass("cfd_volume") {
					selection = selection.Next()
					liters, err := strconv.ParseFloat(selection.Nodes[0].FirstChild.Data, 64)
					if err == nil {
						board.Liters = liters
					}
				} else if selection.HasClass("cfd_size") {
					selection = selection.Next()
					length, err := strconv.ParseFloat(selection.Nodes[0].FirstChild.Data, 64)
					if err == nil {
						board.Length = length
					}
				} else if selection.HasClass("cfd_weight") {
					selection = selection.Next()
					weight, err := strconv.ParseFloat(selection.Nodes[0].FirstChild.Data, 64)
					if err == nil {
						board.Weight = weight
					}
				}
			})

			board.Description = article.ChildText("div.entry-content.subheader")

			fmt.Printf("%+v\n", board)
		})

		c.Visit(e.ChildAttr("a.next.page-numbers", "href"))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	err := c.Visit("https://www.slosurf.com/ad-category/surf/deske-2/")
	if err != nil {
		fmt.Println(err)
		return
	}

}
