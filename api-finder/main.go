package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/gocolly/colly/v2"
)

type api struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	Name       string `json:"name"`
	RawOpenAPI string `json:"rawOpenAPI"`
}

type apiList struct {
	APIs []api `json:"apis"`
}

func (liste *apiList) AddAPI(item api) {
	liste.APIs = append(liste.APIs, item)
}

func (liste *apiList) Sort() {
	// order the list of apis by id
	sort.Slice(liste.APIs, func(i, j int) bool {
		return liste.APIs[i].ID < liste.APIs[j].ID
	})
}

func main() {
	pathFlag := flag.String("path", "geoportal_he.json", "path where to save list of apis")
	geoportalFlag := flag.String("geoportalURL", "https://www.geoportal.hessen.de", "geoportal URL")
	flag.Parse()
	fmt.Println("pathPtr:", *pathFlag)
	fmt.Println("geoportalPtr:", *geoportalFlag)

	list := apiList{}

	// strip https:// from geoportalURL
	geoportalURL := strings.Split(*geoportalFlag, "//")[1]

	c := colly.NewCollector(
		colly.AllowedDomains(geoportalURL),
	)

	cAPI := colly.NewCollector(
		colly.AllowedDomains(geoportalURL),
	)
	cAPI.OnResponse(func(r *colly.Response) {
		// check if content-type header contains JSON
		if strings.Contains(r.Headers.Get("Content-Type"), "json") {
			// read out openapi json
			jsonData := r.Body
			list = extractInfos(jsonData, list)
		}

	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "spatial-objects") {
			// fmt.Printf("Link found: %q -> %s\n", e.Text, link)
			// add /api to link and check if openapi is available
			cAPI.Visit(e.Request.AbsoluteURL(link + "/api"))
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	// Start scraping on geoportalFlag + /spatial-objects
	fmt.Println("Start scraping on", *geoportalFlag+"/spatial-objects/")
	c.Visit(*geoportalFlag + "/spatial-objects/")

	fmt.Println("APIs found:", len(list.APIs))

	list.Sort()

	save2File(list)

}

// save list to json file
func save2File(list apiList) {
	file, _ := json.MarshalIndent(list.APIs, "", " ")

	_ = ioutil.WriteFile("../geoportal_he_go.json", file, 0644)
}

// extract infos from openapi json
func extractInfos(jsonData []byte, list apiList) apiList {

	//parse openapi json
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonData), &jsonMap)
	if err != nil {
		panic(err)
	}

	name := jsonMap["info"].(map[string]interface{})["title"]
	nameStr := fmt.Sprint(name)
	fmt.Println("API name:", nameStr)

	server := jsonMap["servers"].([]interface{})[0].(map[string]interface{})["url"]
	serverStr := fmt.Sprint(server)

	// get id of api (last part of server-url)
	id := strings.Split(serverStr, "/")[4]
	rawOpenAPI := "https://raw.githubusercontent.com/t-huyeng/geoportal-openapis/main/geoportal-he/" + id + ".json"

	// save the data as object to the api_list
	api := api{ID: id, URL: serverStr, Name: nameStr, RawOpenAPI: rawOpenAPI}

	// add api-dict to api_list
	list.AddAPI(api)
	return list
}
