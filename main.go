package main

import (
	"KitsuCrawler/core"
	"fmt"
	"time"
)

func main() {
	start := time.Now()

	userAgent := "KitsuBot"
	websiteToCrawl := "https://www.mbit.pt"

	crawler, err := core.NewCrawler(websiteToCrawl, userAgent)
	if err != nil {
		panic(err.Error())
	}

	crawler.Start()

	elapsed := time.Since(start)

	fmt.Printf("\n\nCrawler took %s", elapsed)
}
