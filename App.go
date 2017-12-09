package main

import (
	"os"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"github.com/jackdanger/collectlinks"
)

var httpClient http.Client

func main() {

	var link = flag.String("url", "", "Url to crawl for dead links")
	flag.Parse()

	baseUrl,  err := url.Parse(*link)
	if err != nil {
		fmt.Printf("Invalid url parameter: %s  : %s", *link, err)
		os.Exit(1);
	}

	deadlinks := make(map[string]error)
	visited := make(map[string]bool)

	traverse (*baseUrl, &visited, &deadlinks, baseUrl, "-")

	fmt.Println("\nDeadlinks found:")
	for k, v := range deadlinks {
		fmt.Printf("- on page %s --> %s\n", k,v)
	}
}

func traverse (url url.URL, visited *map[string]bool, deadlinks *map[string]error, baseUrl *url.URL, parent string) {
	
	if (*visited)[url.String()] {
		return;
	}

	(*visited)[url.String()] = true

	if strings.HasPrefix(url.String(),"mailto") || strings.HasPrefix(url.String(),"javascript") {
		return;
	}
	
	links :=fetch (url ,  deadlinks,  strings.EqualFold(url.Host, baseUrl.Host), parent);

	for _, link := range  links {
		childUrl,  err := url.Parse(link)
	
		if err != nil {
			addDeadlink(deadlinks, url.String(), err)
		}

		resolvedChildUrl := baseUrl.ResolveReference(childUrl)
		
		if !(*visited)[(*resolvedChildUrl).String()] {
			traverse (*resolvedChildUrl, visited, deadlinks, baseUrl, url.String())	
		}
	}
}

func addDeadlink (deadlinks *map[string]error, onPage string, err error) {
	fmt.Printf(`Url: %s Error: %s\n`, onPage, err)
	(*deadlinks)[onPage] = err
}

func  fetch (validUrl url.URL, deadlinks *map[string]error, docollect bool, parent string) []string{

	fmt.Printf("Fetching: %s\n", validUrl.String())
	resp, err := httpClient.Get(validUrl.String())
	if err != nil {
		addDeadlink(deadlinks, parent, err)
		return nil
	}
	defer resp.Body.Close()

	if !docollect {
		//do not check links on third party sites
		return nil;	
	}
	return collectlinks.All(resp.Body)
}
