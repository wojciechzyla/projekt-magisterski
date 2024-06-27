package main

import (
	"fmt"
	ga "genetic_algorithm"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type incomingData struct {
	PoiList  []ga.ApiPOI `json:"poiList"`
	Days     []string    `json:"days"`
	DayStart string      `json:"dayStart"`
	DayEnd   string      `json:"dayEnd"`
}

func getBestRoute(context *gin.Context) {
	var ind incomingData
	if err := context.BindJSON(&ind); err != nil {
		return
	}
	layout := "15:04"
	dayStart, _ := time.Parse(layout, ind.DayStart)
	dayEnd, _ := time.Parse(layout, ind.DayEnd)
	if dayEnd.Before(dayStart) {
		dayEnd = dayEnd.Add(24 * time.Hour)
	}
	dayCode := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	geneticAlgorithm := ga.CreateGeneticAlgorithm(dayStart, dayEnd, ind.Days, 0.05, 1000.0, 1)
	var closingHours map[string]time.Time
	var openingHours map[string]time.Time

	fmt.Println(dayStart)
	fmt.Println(dayEnd)

	for _, p := range ind.PoiList {
		closingHours = make(map[string]time.Time)
		openingHours = make(map[string]time.Time)
		for _, day := range dayCode {
			openingHours[day], _ = time.Parse(layout, p.OpenHour[day])
			closingHours[day], _ = time.Parse(layout, p.CloseHour[day])
			if closingHours[day].Before(openingHours[day]) {
				closingHours[day] = closingHours[day].Add(24 * time.Hour)
			}
		}
		geneticAlgorithm.AddPoi(&ga.POI{
			Name:         p.Name,
			CloseHour:    closingHours,
			OpenHour:     openingHours,
			Lat:          p.Lat,
			Lon:          p.Lon,
			Satisfaction: p.Satisfaction,
		})
	}
	bestItinerary := geneticAlgorithm.Run(300, 100, 8)
	context.IndentedJSON(http.StatusOK, bestItinerary)
}

func main() {
	router := gin.Default()
	router.POST("/best-route", getBestRoute)
	router.Run("0.0.0.0:6000")
}
