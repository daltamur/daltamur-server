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
	cmap "github.com/orcaman/concurrent-map/v2"
	"log"
	"net/http"
	"net/http/pprof"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

//docker run -d -p 33200:8080 daltamur-server

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/daltamur/status", StatusHandler).Methods("GET")
	r.HandleFunc("/daltamur/all", AllHandler).Methods("GET")
	r.HandleFunc("/daltamur/search", RangeHandler).Methods("GET")
	r.HandleFunc("/daltamur/debug/pprof/", pprof.Index)
	r.HandleFunc("/daltamur/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/daltamur/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/daltamur/debug/pprof/symbol", pprof.Symbol)
	r.Handle("/daltamur/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/daltamur/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/daltamur/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/daltamur/debug/pprof/block", pprof.Handler("block"))
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
		going to the default image I set if no artist or track image exists. So a bit of duct tape
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

func convertToDaySongsStruct(output *dynamodb.QueryOutput) DaySongs {
	/*
		Somehow some songs got through with null image references rather than
		going to the default image I set if no artist or track image exists. So a bit of duct tape
		was applied in this function so no nil references get passed and the program breaks
	*/
	var songs DaySongs
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
	sort.Slice(songs.Songs, func(i, j int) bool {
		return songs.Songs[i].UTS > songs.Songs[j].UTS
	})
	return songs
}

func convertToSongsStruct(output *dynamodb.QueryOutput) Songs {
	/*
		Somehow some songs got through with null image references rather than
		going to the default image I set if no artist or track image exists. So a bit of duct tape
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
	monthString := strconv.Itoa(int((*t).Month()))
	if len(monthString) == 1 {
		monthString = "0" + monthString
	}

	dayValString := strconv.Itoa((*t).Day())
	if len(dayValString) == 1 {
		dayValString = "0" + dayValString
	}

	queryInput := dynamodb.QueryInput{
		IndexName: aws.String("date-index"),
		TableName: aws.String("daltamur-LastFMTracks"),
		KeyConditions: map[string]*dynamodb.Condition{
			"date": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(monthString + "/" + dayValString + "/" + year[len(year)-2:]),
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
	jsonBytes = nil
	songs = Songs{}
	//end it off with a loggly message
}

func RangeHandler(writer http.ResponseWriter, request *http.Request) {
	switch len(request.URL.Query()) {
	case 0:
		layout := "01/02/2006 3:04:05 PM"
		location, err := time.LoadLocation("America/New_York")
		var timePointer = new(time.Time)

		month := strconv.Itoa(int(time.Now().In(location).Month()))
		if len(month) == 1 {
			month = "0" + month
		}

		day := strconv.Itoa(int(time.Now().In(location).Day()))
		if len(day) == 1 {
			day = "0" + day
		}

		year := strconv.Itoa(time.Now().In(location).Year())

		*timePointer, err = time.ParseInLocation(layout, month+"/"+day+"/"+year+" 0:00:00 AM", location)
		if err != nil {
			fmt.Println(err)
		}
		//do the filtering here
		fmt.Println("ZONE : ", location, " Time : ", (*timePointer).In(location).Unix())
		filterSingleDay(timePointer, writer)
		msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
		sendLogglyCommand("info", msgVal)

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
					//make sure that the desired date is not greater than the date right now
					//endTime := startTime.AddDate(0, 0, 25)
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
						filterSingleDay(&startTime, writer)
						msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
						sendLogglyCommand("info", msgVal)
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
			errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with wrong query params " + time.Now().String()
			sendLogglyCommand("error", errorVal)
			requestError := Songs{Error: errorVal}
			jsonBytes, _ := json.Marshal(requestError)
			_, err := writer.Write(jsonBytes)
			if err != nil {
				return
			}
		}

	case 2:
		//make sure the start and end date are both defined
		if request.URL.Query().Get("startDate") != "" || request.URL.Query().Get("endDate") != "" {
			//now make sure the start and end date are both well-defined
			r, _ := regexp.Compile("\\b\\d\\d/\\d\\d/\\d\\d\\d\\d\\b")
			if r.MatchString(request.URL.Query().Get("startDate")) && r.MatchString(request.URL.Query().Get("endDate")) {
				slashRegex := regexp.MustCompile("/")
				layout := "01/02/2006 3:04:05 PM"
				location, _ := time.LoadLocation("America/New_York")
				//break the date into the separate parts
				dateArray := slashRegex.Split(request.URL.Query().Get("startDate"), -1)
				startTime, err1 := time.ParseInLocation(layout, dateArray[0]+"/"+dateArray[1]+"/"+dateArray[2]+" 0:00:00 AM", location)
				dateArray = slashRegex.Split(request.URL.Query().Get("endDate"), -1)
				endTime, err2 := time.ParseInLocation(layout, dateArray[0]+"/"+dateArray[1]+"/"+dateArray[2]+" 0:00:00 AM", location)
				endTime = endTime.AddDate(0, 0, 1)
				if err1 != nil || err2 != nil {
					errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with an ill-formed date at " + time.Now().String()
					sendLogglyCommand("error", errorVal)
					requestError := Songs{Error: errorVal}
					jsonBytes, _ := json.Marshal(requestError)
					_, err := writer.Write(jsonBytes)
					if err != nil {
						return
					}
				} else {
					//make sure the start date is not greater than the end date
					if startTime.Unix() > endTime.Unix() || startTime.Unix() > time.Now().In(location).Unix() || endTime.Unix() > time.Now().In(location).Unix() {
						errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with invalid date at " + time.Now().String()
						sendLogglyCommand("error", errorVal)
						requestError := Songs{Error: errorVal}
						jsonBytes, _ := json.Marshal(requestError)
						_, err := writer.Write(jsonBytes)
						if err != nil {
							return
						}
					} else {
						//now do the call
						filterTwoDays(&startTime, &endTime, writer)
						msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
						sendLogglyCommand("info", msgVal)
					}
				}
			} else {
				errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with invalid date format " + time.Now().String()
				sendLogglyCommand("error", errorVal)
				requestError := Songs{Error: errorVal}
				jsonBytes, _ := json.Marshal(requestError)
				_, err := writer.Write(jsonBytes)
				if err != nil {
					return
				}
			}

		} else {
			errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with wrong query params " + time.Now().String()
			sendLogglyCommand("error", errorVal)
			requestError := Songs{Error: errorVal}
			jsonBytes, _ := json.Marshal(requestError)
			_, err := writer.Write(jsonBytes)
			if err != nil {
				return
			}
		}

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

type DirRange []int64

func (a DirRange) Len() int           { return len(a) }
func (a DirRange) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DirRange) Less(i, j int) bool { return a[i] < a[j] }

/*
Using the filter function from DynamoDB is just way too slow to get quality results. Instead, we're going to scan every day in parallel and put them in a concurrent map

*/

func filterTwoDays(t *time.Time, t2 *time.Time, writer http.ResponseWriter) {
	//set the db and filter up
	var wg sync.WaitGroup
	var currentDay *time.Time
	currentDay = t
	dayMap := cmap.New[DaySongs]()
	for !(*currentDay).Equal(*t2) {
		wg.Add(1)
		monthString := strconv.Itoa(int((*currentDay).Month()))
		if len(monthString) == 1 {
			monthString = "0" + monthString
		}

		dayValString := strconv.Itoa((*currentDay).Day())
		if len(dayValString) == 1 {
			dayValString = "0" + dayValString
		}

		yearString := strconv.Itoa((*currentDay).Year())
		var curDayVal = dayValString
		var curMonthVal = monthString
		var curYear = yearString
		var dayValue = *currentDay
		go func() {
			defer wg.Done()
			//fmt.Println(curMonthVal + "/" + curDayVal + "/" + curYear)
			dayMap.Set(curMonthVal+"/"+curDayVal+"/"+curYear, getSingleDayVals(dayValue))
		}()
		*currentDay = (*currentDay).AddDate(0, 0, 1)
	}
	wg.Wait()
	allSongs := SongRange{AllSongs: dayMap.Items()}

	keys := make([]string, 0, len(allSongs.AllSongs))

	for k := range allSongs.AllSongs {
		keys = append(keys, k)
	}

	for _, k := range keys {
		if (allSongs.AllSongs[k]).Songs == nil {
			delete(allSongs.AllSongs, k)
		}
	}

	jsonBytes, _ := json.Marshal(allSongs)

	_, err := writer.Write(jsonBytes)
	if err != nil {
		return
	}
	jsonBytes = nil
	keys = nil
	dayMap = nil
	allSongs.AllSongs = nil
	allSongs = SongRange{}
}

func getSingleDayVals(t time.Time) DaySongs {
	svc := setDBInstance()
	year := strconv.Itoa((t).Year())
	monthString := strconv.Itoa(int((t).Month()))
	if len(monthString) == 1 {
		monthString = "0" + monthString
	}

	dayValString := strconv.Itoa((t).Day())
	if len(dayValString) == 1 {
		dayValString = "0" + dayValString
	}
	queryInput := dynamodb.QueryInput{
		IndexName: aws.String("date-index"),
		TableName: aws.String("daltamur-LastFMTracks"),
		KeyConditions: map[string]*dynamodb.Condition{
			"date": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(monthString + "/" + dayValString + "/" + year[len(year)-2:]),
					},
				},
			},
		},
	}

	returnedVal, _ := svc.Query(&queryInput)
	songs := convertToDaySongsStruct(returnedVal)
	returnedVal.Count = nil
	returnedVal.ConsumedCapacity = nil
	returnedVal.Items = nil
	returnedVal.LastEvaluatedKey = nil
	returnedVal.ScannedCount = nil
	returnedVal = nil
	return songs
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
	errorVal := "404 Error: Page Index " + index + " is not an integer."
	sendLogglyCommand("error", errorVal)
	bytes, _ := json.Marshal(Songs{Error: errorVal})
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
	if int64(page) > numOfPages || int64(page) < 0 {
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
					songs = Songs{}
					jsonBytes = nil
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
	switch len(request.URL.Query()) {
	case 0:
		sendTableData(writer)
		msgVal := "200: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " at " + time.Now().String()
		sendLogglyCommand("info", msgVal)

	default:
		errorVal := "404 Error: " + request.RemoteAddr + " used " + request.Method + " on path " + request.RequestURI + " with too many query params " + time.Now().String()
		writer.Write([]byte(errorVal))
		sendLogglyCommand("error", errorVal)

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
