package main

import (
	"fmt"
	"log"

	"github.com/wispberry-technologies/wispy-core/common"
)

func test_function() {
	siteManager := common.NewSiteManager("./sites")
	site, err := siteManager.LoadSite("localhost")
	if err != nil {
		log.Fatalf("Failed to load site: %v", err)
	}

	pageManager := common.NewPageManager(site)
	pages, err := pageManager.ListPages(true) // Include unpublished pages
	if err != nil {
		log.Fatalf("Failed to list pages: %v", err)
	}

	fmt.Printf("Found %d pages:\n", len(pages))
	for _, page := range pages {
		fmt.Printf("- %s (%s) - Static: %t, URL: %s\n",
			page.Meta.Title, page.Slug, page.Meta.IsStatic, page.Meta.URL)
	}

	indexPage, err := pageManager.GetPage("index")
	if err != nil {
		log.Fatalf("Failed to get index page: %v", err)
	}

	fmt.Printf("\nIndex page details:\n")
	fmt.Printf("Title: %s\n", indexPage.Meta.Title)
	fmt.Printf("Description: %s\n", indexPage.Meta.Description)
	fmt.Printf("Keywords: %v\n", indexPage.Meta.Keywords)
	fmt.Printf("Content length: %d characters\n", len(indexPage.Content))
}
