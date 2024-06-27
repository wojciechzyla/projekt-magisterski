package genetic_algorithm

import (
	"fmt"
	"time"
)

type POI struct {
	Lon          float64              `json:"lon"`
	Lat          float64              `json:"lat"`
	Name         string               `json:"name"`
	OpenHour     map[string]time.Time `json:"openHour"`
	CloseHour    map[string]time.Time `json:"closeHour"`
	Satisfaction float64              `json:"satisfaction"`
}

type ApiPOI struct {
	Lon          float64           `json:"lon"`
	Lat          float64           `json:"lat"`
	Name         string            `json:"name"`
	OpenHour     map[string]string `json:"openHour"`
	CloseHour    map[string]string `json:"closeHour"`
	Satisfaction float64           `json:"satisfaction"`
}

func (poi *POI) print() string {
	return fmt.Sprintf("Name: %s, Lat: %f, Lon: %f, Open: %s, Close: %s, Satisfaction: %f",
		poi.Name, poi.Lat, poi.Lon, poi.OpenHour, poi.CloseHour, poi.Satisfaction)
}

type Visit struct {
	Poi           *POI
	StartVisit    time.Time
	EndVisit      time.Time
	VisitDuration int
}

type ApiVisit struct {
	Poi           ApiPOI
	StartVisit    string
	EndVisit      string
	VisitDuration int
}

type Day struct {
	Visits    []Visit
	DayNumber int
	DayName   string //mon, tue, wed, thu, fri, sat, sun
}

type ApiDay struct {
	Visits    []ApiVisit
	DayNumber int
	DayName   string //mon, tue, wed, thu, fri, sat, sun
}

type Itinerary struct {
	Days         []Day
	DayBeginHour time.Time
	DayEndHour   time.Time
}

type ApiItinerary struct {
	Days         []ApiDay
	DayBeginHour string
	DayEndHour   string
}

func (it *Itinerary) ShortPrint() {
	for _, day := range it.Days {
		fmt.Printf("Day %d:\n", day.DayNumber)
		for _, visit := range day.Visits {
			fmt.Printf(" %s,", visit.Poi.Name)
		}
		fmt.Printf("\n")
	}
}
