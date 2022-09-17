package main

import (
	"encoding/json"
	"fmt"

	"github.com/gocolly/colly"
)

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

func main() {
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

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
	})

	collector.Visit("https://booking.kai.id/?origination=PSE&destination=YK&tanggal=20-September-2022&adult=1&infant=0&submit=Cari+&+Pesan+Tiket")

	jsonData, err := json.MarshalIndent(allTrains, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}
