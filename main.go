package main

import (
	"KitsuCrawler/crawler"
	"fmt"
	"time"
)

func main() {
	start := time.Now()

	userAgent := "KitsuBot"
	websiteToCrawl := "https://www.mbit.pt"

	c, err := crawler.NewCrawler(websiteToCrawl, userAgent)
	if err != nil {
		panic(err.Error())
	}

	c.Start()

	elapsed := time.Since(start)

	fmt.Printf("\n\nCrawler took %s", elapsed)
}
