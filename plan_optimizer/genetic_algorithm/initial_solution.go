package genetic_algorithm

import (
	"math/rand"
	"time"
)

func GenerateRandomItinerary(allPoiList []*POI, dayStart time.Time, dayFinish time.Time, daysList []string) Itinerary {
	rand.Seed(time.Now().UnixNano())
	var startVisit time.Time
	var endVisit time.Time

	usedPoiList := make([]*POI, 0)

	// Assume dayStart and dayFinish are given in the format "HH:mm"
	dayBeginHour := dayStart
	dayEndHour := dayFinish

	itinerary := Itinerary{
		Days:         make([]Day, 0),
		DayBeginHour: dayBeginHour,
		DayEndHour:   dayEndHour,
	}

	for dayNumber, dayName := range daysList {
		visits := make([]Visit, 0)
		//notAvailablePois := make([]*POI, 0)
		prevVisit := (*Visit)(nil)
		poiForDay := make([]*POI, 0)

		// Get a list of available POIs for the day
		for _, poi := range allPoiList {
			if !containsPoi(usedPoiList, poi) {
				poiForDay = append(poiForDay, poi)
			}
		}

		for {
			// Break if no more available POIs or end time is reached
			if len(poiForDay) == 0 || (prevVisit != nil && prevVisit.EndVisit.After(dayEndHour)) || (prevVisit != nil && addMinutes(prevVisit.EndVisit, 60).After(dayEndHour)) {
				break
			}

			newPoi, newPoiIndex := drawPoi(poiForDay)

			// Calculate start and end times for the new POI
			if prevVisit != nil {
				startVisit = prevVisit.EndVisit
				startVisit = addMinutes(startVisit, transport(prevVisit.Poi, newPoi))
			} else {
				startVisit = dayBeginHour
			}
			if newPoi.OpenHour[dayName].After(startVisit) {
				timeDiff := calculateDuration(startVisit, newPoi.OpenHour[dayName])
				startVisit = addMinutes(startVisit, timeDiff)
			}
			endVisit = minHour(addMinutes(startVisit, rand.Intn(121)+60), newPoi.CloseHour[dayName], dayEndHour)
			if startVisit.After(endVisit) || startVisit.Equal(endVisit) || calculateDuration(startVisit, endVisit) < 60 {
				poiForDay[newPoiIndex] = poiForDay[len(poiForDay)-1]
				poiForDay = poiForDay[:len(poiForDay)-1]
				continue
			}

			visit := Visit{
				Poi:           newPoi,
				StartVisit:    startVisit,
				EndVisit:      endVisit,
				VisitDuration: calculateDuration(startVisit, endVisit),
			}

			visits = append(visits, visit)
			usedPoiList = append(usedPoiList, newPoi)
			poiForDay[newPoiIndex] = poiForDay[len(poiForDay)-1]
			poiForDay = poiForDay[:len(poiForDay)-1]
			prevVisit = &visit
		}

		day := Day{
			Visits:    visits,
			DayNumber: dayNumber,
			DayName:   dayName,
		}

		itinerary.Days = append(itinerary.Days, day)
	}
	return itinerary
}
