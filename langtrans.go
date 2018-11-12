package main

import (
	"fmt"
	"os"
	"log"
	"net/http"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	
	//core "github.com/watson-developer-cloud/go-sdk/core"
	languagetranslator "github.com/watson-developer-cloud/go-sdk/languagetranslatorv3"
)

import "github.com/timjacobi/go-couchdb"


type Visitor struct {
	Name string `json:"name"`
}

type Visitors []Visitor

type alldocsResult struct {
	TotalRows int `json:"total_rows"`
	Offset    int
	Rows      []map[string]interface{}
}


func afterxxx(value string, a string) string {
    // Get substring after a string.
    pos := strings.LastIndex(value, a)
    if pos == -1 {
        return ""
    }
    adjustedPos := pos + len(a)
    if adjustedPos >= len(value) {
        return ""
    }
    return value[adjustedPos:len(value)]
}



func main() {




//database
r := gin.Default()

	r.StaticFile("/", "./static/index.html")
	r.Static("/static", "./static")

	var dbName = "mydb"

	//When running locally, get credentials from .env file.
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file does not exist")
	}
	cloudantUrl := os.Getenv("CLOUDANT_URL")

	appEnv, _ := cfenv.Current()
	if appEnv != nil {
		cloudantService, _ := appEnv.Services.WithLabel("cloudantNoSQLDB")
		if len(cloudantService) > 0 {
			cloudantUrl = cloudantService[0].Credentials["url"].(string)
		}
	}
	
	cloudant, err := couchdb.NewClient(cloudantUrl, nil)
	if err != nil {
		log.Println("Can not connect to Cloudant database")
	}

	//ensure db exists
	//if the db exists the db will be returned anyway
	cloudant.CreateDB(dbName)
	
	
	




	// Instantiate the Watson Language Translator service
	service, serviceErr := languagetranslator.
		NewLanguageTranslatorV3(&languagetranslator.LanguageTranslatorV3Options{
			URL:       "https://gateway-wdc.watsonplatform.net/language-translator/api",
			Version:   "2018-11-01",
			IAMApiKey: "IG3K7VsHI4lEuQUcKnw4rek_aSnvhR2DZK2JTccteetJ",
		})

	// Check successful instantiation
	if serviceErr != nil {
		fmt.Println(serviceErr)
		return
	}

	/* TRANSLATE */

	textToTranslate := []string{
		"you are love",
		"too much love",
	}

	translateOptions := service.NewTranslateOptions(textToTranslate).
		SetModelID("en-hi")

	// Call the languageTranslator Translate method
	response, responseErr := service.Translate(translateOptions)

	// Check successful call
	if responseErr != nil {
		panic(responseErr)
	}

	//fmt.Println(response)
	
	 finalstring:=afterxxx(response.String(), "\"translation\": \"")
	 finalstring=finalstring[:len(finalstring)-28]
    fmt.Println(strings.Split(finalstring,"\"")[0])



		
	
	
	r.POST("/api/visitors", func(c *gin.Context) {
		var visitor Visitor
		if c.BindJSON(&visitor) == nil {
			cloudant.DB(dbName).Post(visitor)
			c.String(200, finalstring)
		}
	})

 



	
	
	r.GET("/api/visitors", func(c *gin.Context) {
		var result alldocsResult
		if cloudantUrl == "" {
			c.JSON(200, gin.H{})
			return
		}
		err := cloudant.DB(dbName).AllDocs(&result, couchdb.Options{"include_docs": true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch docs"})
		} else {
			c.JSON(200, result.Rows)
		}
	})
	
		

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" //Local
	}
	r.Run(":" + port)
}