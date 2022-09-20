package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
)

var (
	ErrorInvalidQueryParameters = "Missing or invalid query paramenter(s)"
	ErrorNoDataFound            = "No data for the requested route"
	MonthsData                  = map[string]string{
		"01": "Januari",
		"02": "Februari",
		"03": "Maret",
		"04": "April",
		"05": "Mei",
		"06": "Juni",
		"07": "Juli",
		"08": "Agustus",
		"09": "September",
		"10": "Oktober",
		"11": "November",
		"12": "Desember",
	}
)

type ErrorBody struct {
	Message string `json:"message"`
}

type TrainOrigin struct {
	DepartureStation string `json:"departureStation"`
	DepartureDate    string `json:"departureDate"`
	DepartureTime    string `json:"departureTime"`
}

type TrainDestination struct {
	ArrivalStation string `json:"arrivalStation"`
	ArrivalDate    string `json:"arrivalDate"`
	ArrivalTime    string `json:"arrivalTime"`
}

type Train struct {
	TrainName        string           `json:"trainName"`
	TrainClass       string           `json:"trainClass"`
	TrainOrigin      TrainOrigin      `json:"trainOrigin"`
	TrainDestination TrainDestination `json:"trainDestination"`
	TravelTime       string           `json:"travelTime"`
	TicketPrice      string           `json:"ticketPrice"`
	SeatAvailability string           `json:"seatAvailability"`
}

func isStationValid(station string) bool {
	var rxStation = regexp.MustCompile(`[A-Z]{2,3}`)

	return rxStation.MatchString(station)
}

func isDateValid(date string) bool {
	var rxDate = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

	return rxDate.MatchString(date)
}

func transformDate(date string) string {
	dateArray := strings.Split(date, "-")
	dateArray[1] = MonthsData[dateArray[1]]

	for i := len(dateArray)/2 - 1; i >= 0; i-- {
		opp := len(dateArray) - 1 - i
		dateArray[i], dateArray[opp] = dateArray[opp], dateArray[i]
	}

	newDate := strings.Join(dateArray, "-")

	return newDate
}

func scraper(url string) []Train {
	allTrains := make([]Train, 0)

	collector := colly.NewCollector(
		colly.AllowedDomains("booking.kai.id"),
	)

	collector.OnHTML("div.data-wrapper", func(element *colly.HTMLElement) {
		trainName := element.ChildText("div.name")
		trainClass := element.DOM.Find("div.col-one").Children().Last().Text()

		departureStation := element.ChildText("div.station-start")
		departureDate := element.ChildText("div.date-start")
		departureTime := element.ChildText("div.time-start")

		origin := TrainOrigin{
			DepartureStation: departureStation,
			DepartureDate:    departureDate,
			DepartureTime:    departureTime,
		}

		arrivalStation := element.DOM.Find("div.card-arrival").Children().First().Text()
		arrivalDate := element.DOM.Find("div.card-arrival").Children().Last().Text()
		arrivalTime := element.ChildText("div.time-end")

		destination := TrainDestination{
			ArrivalStation: arrivalStation,
			ArrivalDate:    arrivalDate,
			ArrivalTime:    arrivalTime,
		}

		travelTime := element.ChildText("div.long-time")
		ticketPrice := element.ChildText("div.price")
		seatAvailability := element.ChildText("small.sisa-kursi")

		train := Train{
			TrainName:        trainName,
			TrainClass:       trainClass,
			TrainOrigin:      origin,
			TrainDestination: destination,
			TravelTime:       travelTime,
			TicketPrice:      ticketPrice,
			SeatAvailability: seatAvailability,
		}

		allTrains = append(allTrains, train)
	})

	collector.Visit(url)

	return allTrains
}

func apiResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	res := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	res.StatusCode = statusCode
	jsonBody, _ := json.MarshalIndent(body, "", "  ")
	res.Body = string(jsonBody)

	return res, nil
}

func lambdaHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	from := req.QueryStringParameters["from"]
	to := req.QueryStringParameters["to"]
	date := req.QueryStringParameters["date"]

	if !isStationValid(from) || !isStationValid(to) || !isDateValid(date) {
		return apiResponse(http.StatusBadRequest, ErrorBody{Message: ErrorInvalidQueryParameters})
	}

	urlToScrape, _ := url.Parse("https://booking.kai.id/")
	query := urlToScrape.Query()
	query.Set("origination", from)
	query.Set("destination", to)
	query.Set("tanggal", transformDate(date))
	query.Set("adult", "1")
	query.Set("infant", "0")
	query.Set("submit", "Cari+&+Pesan+Tiket")
	urlToScrape.RawQuery = query.Encode()
	result := scraper(urlToScrape.String())

	if len(result) == 0 {
		return apiResponse(http.StatusNotFound, ErrorBody{Message: ErrorNoDataFound})
	}

	return apiResponse(http.StatusOK, result)
}

func main() {
	lambda.Start(lambdaHandler)
}
