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
	saveFlag := flag.Bool("save", false, "save openapis	to file")
	flag.Parse()
	fmt.Println("pathFlag:", *pathFlag)
	fmt.Println("geoportalFlag:", *geoportalFlag)
	fmt.Println("saveFlag:", *saveFlag)

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
			list = extractInfos(r, list, *saveFlag)
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
	// Start scraping on geoportalFlag + /spatial-objects/
	fmt.Println("Start scraping on", *geoportalFlag+"/spatial-objects/")
	c.Visit(*geoportalFlag + "/spatial-objects/")

	fmt.Println("APIs found:", len(list.APIs))

	list.Sort()

	save2File(list, pathFlag)

}

// save list to json file
func save2File(list apiList, pathFlag *string) {

	fmt.Println("Saving API list to file...")
	file, _ := json.MarshalIndent(list.APIs, "", " ")

	_ = ioutil.WriteFile(*pathFlag, file, 0644)
}

// extract infos from openapi json
func extractInfos(r *colly.Response, list apiList, save bool) apiList {

	jsonData := r.Body

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
	rawOpenAPI := ""
	// call URL and check HEADERS for CORS
	if strings.Contains((r.Headers.Get("Access-Control-Allow-Origin")), "*") {
		fmt.Println("- CORS allowed")
		// check if serverStr is ending with trailing slash
		if strings.HasSuffix(serverStr, "/") {
			rawOpenAPI = serverStr + "api"
		} else {
			rawOpenAPI = serverStr + "/api"
		}
	} else {
		fmt.Println("- CORS not allowed")
		folder := "unknown"
		// check URL and set folder for rawOpenAPI
		if strings.Contains(serverStr, "geoportal.hessen.de") {
			folder = "geoportal-he"
		} else if strings.Contains(serverStr, "geoportal.saarland.de") {
			folder = "geoportal-sl"
		} else if strings.Contains(serverStr, "geoportal.rlp.de") {
			folder = "geoportal-rp"
		}

		// if save flag is active save openapi to file into the folder
		if save {
			fmt.Println("Saving OpenAPI to file..." + folder + "/" + id + ".json")
			err = ioutil.WriteFile("../"+folder+"/"+id+".json", r.Body, 0644)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				rawOpenAPI = "https://raw.githubusercontent.com/t-huyeng/geoportal-openapis/main/" + folder + "/" + id + ".json"
			}
		}
	}
	fmt.Println("Raw OpenAPI:", rawOpenAPI)
	// save the data as object to the api_list
	api := api{ID: id, URL: serverStr, Name: nameStr, RawOpenAPI: rawOpenAPI}

	// add api-dict to api_list
	list.AddAPI(api)
	return list
}
