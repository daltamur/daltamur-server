package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"
)

//docker run -d -p 33200:8080 daltamur-server

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/daltamur/status", StatusHandler).Methods("GET")
	r.HandleFunc("/daltamur/all", AllHandler).Methods("GET")
	r.HandleFunc("/daltamur/select", RangeHandler).Methods("GET")
	r.HandleFunc("/{path:.+}", ErrorHandler)
	r.Methods("POST", "PUT", "PATCH", "DELETE").HandlerFunc(ErrorHandler)
	http.Handle("/", r)

	acceptedMethods := handlers.AllowedMethods([]string{"GET"})
	srv := &http.Server{
		Handler:      handlers.CORS(acceptedMethods)(r),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func convertToSongsStructScan(output *dynamodb.ScanOutput) Songs {
	/*
		Somehow some songs got through with null image references rather than
		going to the default image I set if no artist or track image exists. So a little bit of ducttape
		was applied in this function so no nil references get passed and the program breaks
	*/
	var songs Songs
	for i := range output.Items {
		var images []Image
		for x := range output.Items[i]["artist-image"].L {
			var curImage Image
			if ((*output).Items[i]["artist-image"].L[x].M["size"].NULL) == nil {
				curImage.Size = *((*output).Items[i]["artist-image"].L[x].M["size"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["artist-image"].L[x].M["size"].NULL) != true {
					curImage.Size = *((*output).Items[i]["artist-image"].L[x].M["size"].S)
				}
			}

			if ((*output).Items[i]["artist-image"].L[x].M["#text"].NULL) == nil {
				curImage.Text = *((*output).Items[i]["artist-image"].L[x].M["#text"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["artist-image"].L[x].M["#text"].NULL) != true {
					curImage.Text = *((*output).Items[i]["artist-image"].L[x].M["size"].S)
				}
			}
			images = append(images, curImage)
		}

		var albumImages []Image

		for x := range output.Items[i]["track-image"].L {
			var curImage Image
			if ((*output).Items[i]["track-image"].L[x].M["size"].NULL) == nil {
				curImage.Size = *((*output).Items[i]["track-image"].L[x].M["size"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["track-image"].L[x].M["size"].NULL) != true {
					curImage.Size = *((*output).Items[i]["track-image"].L[x].M["size"].S)
				}
			}

			if ((*output).Items[i]["track-image"].L[x].M["#text"].NULL) == nil {
				curImage.Text = *((*output).Items[i]["track-image"].L[x].M["#text"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["track-image"].L[x].M["#text"].NULL) != true {
					curImage.Text = *((*output).Items[i]["track-image"].L[x].M["size"].S)
				}
			}
			albumImages = append(albumImages, curImage)
		}

		uts, _ := strconv.Atoi(*output.Items[i]["uts-time"].N)

		songStruct := SongData{
			Artist:       *output.Items[i]["artist"].S,
			Album:        *output.Items[i]["album"].S,
			ArtistImages: images,
			AlbumImages:  albumImages,
			Date:         *output.Items[i]["date"].S,
			Name:         *output.Items[i]["track"].S,
			Time:         *output.Items[i]["EST-time"].S,
			UTS:          uts,
		}

		songs.Songs = append(songs.Songs, songStruct)
	}
	songs.CurPage = "0"
	sort.Slice(songs.Songs, func(i, j int) bool {
		return songs.Songs[i].UTS > songs.Songs[j].UTS
	})
	return songs
}

func convertToSongsStruct(output *dynamodb.QueryOutput) Songs {
	/*
		Somehow some songs got through with null image references rather than
		going to the default image I set if no artist or track image exists. So a little bit of ducttape
		was applied in this function so no nil references get passed and the program breaks
	*/
	var songs Songs
	for i := range output.Items {
		var images []Image
		for x := range output.Items[i]["artist-image"].L {
			var curImage Image
			if ((*output).Items[i]["artist-image"].L[x].M["size"].NULL) == nil {
				curImage.Size = *((*output).Items[i]["artist-image"].L[x].M["size"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["artist-image"].L[x].M["size"].NULL) != true {
					curImage.Size = *((*output).Items[i]["artist-image"].L[x].M["size"].S)
				}
			}

			if ((*output).Items[i]["artist-image"].L[x].M["#text"].NULL) == nil {
				curImage.Text = *((*output).Items[i]["artist-image"].L[x].M["#text"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["artist-image"].L[x].M["#text"].NULL) != true {
					curImage.Text = *((*output).Items[i]["artist-image"].L[x].M["size"].S)
				}
			}
			images = append(images, curImage)
		}

		var albumImages []Image

		for x := range output.Items[i]["track-image"].L {
			var curImage Image
			if ((*output).Items[i]["track-image"].L[x].M["size"].NULL) == nil {
				curImage.Size = *((*output).Items[i]["track-image"].L[x].M["size"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["track-image"].L[x].M["size"].NULL) != true {
					curImage.Size = *((*output).Items[i]["track-image"].L[x].M["size"].S)
				}
			}

			if ((*output).Items[i]["track-image"].L[x].M["#text"].NULL) == nil {
				curImage.Text = *((*output).Items[i]["track-image"].L[x].M["#text"].S)
			} else {
				//one final check to make sure it isn't true
				if *((*output).Items[i]["track-image"].L[x].M["#text"].NULL) != true {
					curImage.Text = *((*output).Items[i]["track-image"].L[x].M["size"].S)
				}
			}
			albumImages = append(albumImages, curImage)
		}

		uts, _ := strconv.Atoi(*output.Items[i]["uts-time"].N)

		songStruct := SongData{
			Artist:       *output.Items[i]["artist"].S,
			Album:        *output.Items[i]["album"].S,
			ArtistImages: images,
			AlbumImages:  albumImages,
			Date:         *output.Items[i]["date"].S,
			Name:         *output.Items[i]["track"].S,
			Time:         *output.Items[i]["EST-time"].S,
			UTS:          uts,
		}

		songs.Songs = append(songs.Songs, songStruct)
	}
	songs.CurPage = "0"
	sort.Slice(songs.Songs, func(i, j int) bool {
		return songs.Songs[i].UTS > songs.Songs[j].UTS
	})
	return songs
}

func filterSingleDay(t *time.Time, writer http.ResponseWriter) {
	svc := setDBInstance()
	year := strconv.Itoa((*t).Year())
	queryInput := dynamodb.QueryInput{
		IndexName: aws.String("date-index"),
		TableName: aws.String("daltamur-LastFMTracks"),
		KeyConditions: map[string]*dynamodb.Condition{
			"date": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(strconv.Itoa(int((*t).Month())) + "/" + strconv.Itoa((*t).Day()) + "/" + year[len(year)-2:]),
					},
				},
			},
		},
	}

	returnedVal, _ := svc.Query(&queryInput)
	songs := convertToSongsStruct(returnedVal)
	jsonBytes, _ := json.Marshal(songs)
	_, err := writer.Write(jsonBytes)
	if err != nil {
		return
	}

	//end it off with a loggly message
}

func RangeHandler(writer http.ResponseWriter, request *http.Request) {
	switch len(request.URL.Query()) {
	case 0:
		layout := "01/02/2006 3:04:05 PM"
		location, err := time.LoadLocation("America/New_York")
		var timePointer = new(time.Time)
		*timePointer, err = time.ParseInLocation(layout, strconv.Itoa(int(time.Now().In(location).Month()))+"/"+strconv.Itoa(time.Now().In(location).Day())+"/"+strconv.Itoa(time.Now().In(location).Year())+" 0:00:00 AM", location)
		if err != nil {
			fmt.Println(err)
		}
		//do the filtering here
		fmt.Println("ZONE : ", location, " Time : ", (*timePointer).In(location).Unix())
		filterSingleDay(timePointer, writer)

	case 1:
		if request.URL.Query().Get("startDate") != "" {
			//date regex
			r, _ := regexp.Compile("\\b\\d\\d/\\d\\d/\\d\\d\\d\\d\\b")
			//make sure that the user put in a valid pattern before proceeding
			if r.MatchString(request.URL.Query().Get("startDate")) != false {
				slashRegex := regexp.MustCompile("/")
				//break the date into the separate parts
				dateArray := slashRegex.Split(request.URL.Query().Get("startDate"), -1)
				layout := "01/02/2006 3:04:05 PM"
				location, _ := time.LoadLocation("America/New_York")
				startTime, err := time.ParseInLocation(layout, dateArray[0]+"/"+dateArray[1]+"/"+dateArray[2]+" 0:00:00 AM", location)
				//make sure that the date that was given is a valid date
				if err != nil {
					errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with an ill-formed date at " + time.Now().String()
					sendLogglyCommand("error", errorVal)
					requestError := Songs{Error: errorVal}
					jsonBytes, _ := json.Marshal(requestError)
					_, err := writer.Write(jsonBytes)
					if err != nil {
						return
					}
				} else {
					endTime := startTime.AddDate(0, 0, 1)
					//make sure that the desired date is not greater than the date right now
					if startTime.Unix() > time.Now().In(location).Unix() {
						errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with invalid date at " + time.Now().String()
						sendLogglyCommand("error", errorVal)
						requestError := Songs{Error: errorVal}
						jsonBytes, _ := json.Marshal(requestError)
						_, err := writer.Write(jsonBytes)
						if err != nil {
							return
						}
					} else {
						//finally do the scan of the database here
						//do this tmrrw
						fmt.Println(startTime.In(location))
						fmt.Println(endTime.In(location))
					}
				}
			} else {
				errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with invalid date at " + time.Now().String()
				sendLogglyCommand("error", errorVal)
				requestError := Songs{Error: errorVal}
				jsonBytes, _ := json.Marshal(requestError)
				_, err := writer.Write(jsonBytes)
				if err != nil {
					return
				}
			}
		} else {
			errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with wrong query param " + time.Now().String()
			sendLogglyCommand("error", errorVal)
			requestError := Songs{Error: errorVal}
			jsonBytes, _ := json.Marshal(requestError)
			_, err := writer.Write(jsonBytes)
			if err != nil {
				return
			}
		}

	case 2:
		//do this tmrrw too

	default:
		errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with invalid amount of query params at " + time.Now().String()
		sendLogglyCommand("error", errorVal)
		requestError := Songs{Error: errorVal}
		jsonBytes, _ := json.Marshal(requestError)
		_, err := writer.Write(jsonBytes)
		if err != nil {
			return
		}

	}

}

func AllHandler(writer http.ResponseWriter, request *http.Request) {
	var pageNum int
	var err error

	switch len(request.URL.Query()) {
	case 0:
		pageNum = 0
		sendAllTableData(writer, pageNum)
		msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
		sendLogglyCommand("info", msgVal)
	case 1:
		if request.URL.Query().Get("page") == "" {
			errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with undefined query at " + time.Now().String()
			sendLogglyCommand("error", errorVal)
			requestError := Songs{Error: errorVal}
			jsonBytes, _ := json.Marshal(requestError)
			_, err := writer.Write(jsonBytes)
			if err != nil {
				return
			}
		} else {
			pageNum, err = strconv.Atoi(request.URL.Query().Get("page"))
			if err != nil {
				writePageSizeError(writer, request.URL.Query().Get("page"))
			} else {
				sendAllTableData(writer, pageNum)
				msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
				sendLogglyCommand("info", msgVal)
			}
		}
	default:
		errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with too many query params at " + time.Now().String()
		sendLogglyCommand("error", errorVal)
		requestError := Songs{Error: errorVal}
		jsonBytes, _ := json.Marshal(requestError)
		_, err := writer.Write(jsonBytes)
		if err != nil {
			return
		}

	}
}

func writePageSizeError(writer http.ResponseWriter, index string) {
	error := "404 Error: Page Index " + index + " is not an integer."
	sendLogglyCommand("error", error)
	bytes, _ := json.Marshal(Songs{Error: error})
	writer.Write(bytes)
}

func getNumOfPages(svc *dynamodb.DynamoDB) int64 {
	tableDescription, _ := svc.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String("daltamur-LastFMTracks")})
	numOfPages := *tableDescription.Table.ItemCount / int64(200)
	//need an extra page for any possible remainder
	if numOfPages%200 == 0 {
		numOfPages = numOfPages - 1
	}
	return numOfPages

}

func sendAllTableData(writer http.ResponseWriter, page int) {
	foundPage := true
	svc := setDBInstance()
	numOfPages := getNumOfPages(svc)
	//if the requested entry is larger than the number of pages we have
	if int64(page) > numOfPages {
		errorVal := "404 Error: Index " + strconv.Itoa(page) + " out of bounds for length " + strconv.Itoa(int(numOfPages))
		sendLogglyCommand("error", errorVal)
		requestError := Songs{Error: errorVal}
		jsonBytes, _ := json.Marshal(requestError)
		_, err := writer.Write(jsonBytes)
		if err != nil {
			return
		}
	} else {
		pageNum := 0
		scanInput := dynamodb.ScanInput{
			IndexName: aws.String("uts-time-album-index"),
			TableName: aws.String("daltamur-LastFMTracks"),
			Limit:     aws.Int64(200),
		}
		_ = svc.ScanPages(&scanInput,
			func(output *dynamodb.ScanOutput, b bool) bool {
				if pageNum == page {
					pageCount := "Page " + strconv.Itoa(pageNum) + " of " + strconv.Itoa(int(numOfPages))
					var songs Songs
					songs = convertToSongsStructScan(output)
					songs.CurPage = pageCount
					jsonBytes, _ := json.Marshal(songs)
					_, err := writer.Write(jsonBytes)
					if err != nil {
						return false
					}
					foundPage = false
				}
				pageNum++
				return foundPage
			})
	}

}

func ErrorHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("404 Path Not Found."))
	errorVal := "404 Error: " + request.RemoteAddr + " attempted to use " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
	sendLogglyCommand("error", errorVal)
}

func StatusHandler(writer http.ResponseWriter, request *http.Request) {
	sendTableData(writer)
	switch len(request.URL.Query()) {
	case 0:
		msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
		sendLogglyCommand("info", msgVal)

	default:
		msgVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with too many query params at " + time.Now().String()
		sendLogglyCommand("info", msgVal)

	}
}

func sendTableData(writer http.ResponseWriter) {
	svc := setDBInstance()
	tableDescription, _ := svc.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String("daltamur-LastFMTracks")})
	tableStatus := Status{Table: *tableDescription.Table.TableName, RecordCount: *tableDescription.Table.ItemCount, Time: time.Now().String()}
	jsonBytes, _ := json.Marshal(tableStatus)
	_, err := writer.Write(jsonBytes)
	if err != nil {
		return
	}
}

func setDBInstance() *dynamodb.DynamoDB {
	//instantiate the reference to the DB
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	return svc

}

func sendLogglyCommand(msgType string, msg string) {
	logglyClient := loggly.New("Server-Requests")
	err := logglyClient.EchoSend(msgType, msg)
	if err != nil {
		return
	}
}
