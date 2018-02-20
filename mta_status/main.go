package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Response models a return value for AWS Lambda
type Response struct {
	Message string `json:"message"`
}

// Service models the relevant data of MTA's subway status
type Service struct {
	ResponseCode string `xml:"responsecode"`
	TimeStamp    string `xml:"timestamp"`
	Subways      Subway `xml:"subway"`
}

// Subway contains a list of subways by line
type Subway struct {
	Lines []Line `xml:"line"`
}

// Line contains a specific track of 1 or more subways,
// and their current status
type Line struct {
	Name   string `xml:"name"`
	Status string `xml:"status"`
	Date   string `xml:"Date"`
	Time   string `xml:"Time"`
}

/* Handler runs our core functionality for our lambda
*  Requests the most updated subway data from MTA
*  Parses the response into a Service struct
 */
func Handler() (Response, error) {
	mtaData, err := http.Get("http://web.mta.info/status/serviceStatus.txt")
	if err != nil {
		return GetErrorMsg(err)
	}

	defer mtaData.Body.Close()
	body, err := ioutil.ReadAll(mtaData.Body)
	if err != nil {
		return GetErrorMsg(err)
	}

	var mtaDataBody Service
	xml.Unmarshal(body, &mtaDataBody)
	subwayMap := GetDataBySubwayLine(mtaDataBody)
	PrintLinesByStatus(subwayMap)
	return Response{
		Message: string(body),
	}, nil
}

/* GetDataBySubwayLine will parse the individual
*  lines text, and will format the data
*  for Alexa to read from
 */
func GetDataBySubwayLine(mtaData Service) map[string]string {
	var subwayMap = make(map[string]string)

	for _, lines := range mtaData.Subways.Lines {
		MapLineNames(subwayMap, lines)
	}
	return subwayMap
}

/*  MapLinesNames maps the subway status to
    a subway line
*/
func MapLineNames(subwayMap map[string]string, linesList Line) {
	// Staten Island Railroad one exception
	if linesList.Name == "SIR" {
		subwayMap[linesList.Name] = linesList.Status
		return
	}

	for _, line := range linesList.Name {
		subwayMap[string(line)] = linesList.Status
	}
}

/* GetLatestUpdateTime will return the timestamp
   of MTA Service update
*/
func GetLatestUpdateTime(mtaData Service) string {
	return mtaData.TimeStamp
}

/* GetErrorMsg returns a Response with an error msg
*  called whenever something goes wrong and we capture
*  and error
 */
func GetErrorMsg(err error) (Response, error) {
	return Response{
		Message: "Got back err: " + err.Error(),
	}, err
}

/*
	PrintLinesByStatus prints subway statuses by lines
	only used for debugging
*/
func PrintLinesByStatus(subwayMap map[string]string) {
	for k, v := range subwayMap {
		fmt.Println("line: ", k, ": ", v)
	}
}

func main() {
	Handler()
	//lambda.Start(Handler) todo: test locally for now
}
