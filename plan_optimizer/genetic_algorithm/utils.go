package genetic_algorithm

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const radius = 6371 // Earth's mean radius in kilometers

func degrees2radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func transport(prevPoi *POI, newPoi *POI) int {
	degreesLat := degrees2radians(newPoi.Lat - prevPoi.Lat)
	degreesLong := degrees2radians(newPoi.Lon - prevPoi.Lon)
	a := math.Sin(degreesLat/2)*math.Sin(degreesLat/2) +
		math.Cos(degrees2radians(prevPoi.Lat))*
			math.Cos(degrees2radians(newPoi.Lat))*math.Sin(degreesLong/2)*
			math.Sin(degreesLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := radius * c

	return int(math.Ceil(d * 6.0))
}

func drawPoi(poiList []*POI) (*POI, int) {
	// Randomly select a POI from the available list
	if len(poiList) == 0 {
		return nil, -1
	}
	index := rand.Intn(len(poiList))
	return poiList[index], index
}

func CompareHours(timeStr1, timeStr2 string) int {
	layout := "15:04"
	t1, err1 := time.Parse(layout, timeStr1)
	t2, err2 := time.Parse(layout, timeStr2)

	if err1 != nil || err2 != nil {
		fmt.Println("Error parsing time:", err1, err2)
		return 0
	}

	if t1.After(t2) {
		return 1 // timeStr1 is later
	} else if t1.Before(t2) {
		return -1 // timeStr1 is earlier
	}

	return 0 // Both times are equal
}

func addMinutes(startTime time.Time, minutes int) time.Time {
	return startTime.Add(time.Minute * time.Duration(minutes))
}

func subtractMinutes(startTime time.Time, minutes int) time.Time {
	return startTime.Add(-time.Minute * time.Duration(minutes))
}

func calculateDuration(startTime, endTime time.Time) int {
	duration := endTime.Sub(startTime)
	result := int(duration.Minutes())
	return result
}

func containsPoi(poiList []*POI, poi *POI) bool {
	for _, p := range poiList {
		if p == poi {
			return true
		}
	}
	return false
}

func openedDuringHours(poi *POI, startHour, endHour time.Time, day string) bool {
	openTime := poi.OpenHour[day]
	closeTime := poi.CloseHour[day]
	startTime := startHour
	endTime := endHour

	// If closeTime is on or after 00:00, adjust it to the next day
	if closeTime.Before(openTime) {
		closeTime = closeTime.Add(24 * time.Hour)
	}

	// If endTime is on or after 00:00, adjust it to the next day
	if endTime.Before(startTime) {
		endTime = endTime.Add(24 * time.Hour)
	}

	// Check if the opening hours are within the specified start and end hours
	return (startTime.After(openTime) || startTime.Equal(openTime)) &&
		(endTime.Before(closeTime) || endTime.Equal(closeTime))
}

func doesNewPoiFit(newPoi *POI, visits []Visit, visitId int, dayStartHour time.Time, dayEndHour time.Time, day string) (result bool,
	visitStart time.Time, visitEnd time.Time) {
	result = false
	visitStart = time.Time{}
	visitEnd = time.Time{}

	if visitId == 0 {
		// Substitute first point during the day
		if len(visits) > 1 {
			travelTime := transport(newPoi, visits[1].Poi)
			startTime := maxHour(newPoi.OpenHour[day], dayStartHour)
			endTime := minHour(newPoi.CloseHour[day], subtractMinutes(visits[1].StartVisit, travelTime))
			if startTime.Before(endTime) && calculateDuration(startTime, endTime) >= 60 {
				if openedDuringHours(newPoi, startTime, endTime, day) {
					result = true
					visitStart = startTime
					visitEnd = endTime
				}
			}
		} else {
			if openedDuringHours(newPoi, visits[visitId].StartVisit, visits[visitId].EndVisit, day) {
				result = true
				visitStart = visits[visitId].StartVisit
				visitEnd = visits[visitId].EndVisit
			}
		}
	} else if visitId < len(visits)-1 {
		// Substitute point in the middle of the day
		travelFromPrev := transport(visits[visitId-1].Poi, newPoi)
		travelToNext := transport(newPoi, visits[visitId+1].Poi)

		startTime := maxHour(newPoi.OpenHour[day], addMinutes(visits[visitId-1].EndVisit, travelFromPrev))
		endTime := minHour(newPoi.CloseHour[day], subtractMinutes(visits[visitId+1].StartVisit, travelToNext))

		if startTime.Before(endTime) && calculateDuration(startTime, endTime) >= 60 {
			if openedDuringHours(newPoi, startTime, endTime, day) {
				result = true
				visitStart = startTime
				visitEnd = endTime
			}
		}
	} else {
		// Substitute point at the end of the day
		travelTime := transport(visits[visitId-1].Poi, newPoi)
		startTime := maxHour(newPoi.OpenHour[day], addMinutes(visits[visitId-1].EndVisit, travelTime))
		endTime := minHour(newPoi.CloseHour[day], dayEndHour)
		if startTime.Before(endTime) && calculateDuration(startTime, endTime) >= 60 {
			if openedDuringHours(newPoi, startTime, endTime, day) {
				result = true
				visitStart = startTime
				visitEnd = endTime
			}
		}
	}
	return result, visitStart, visitEnd
}

func minHour(times ...time.Time) time.Time {
	if len(times) == 0 {
		// No arguments provided, return a zero time
		return time.Time{}
	}

	min := times[0]
	for _, t := range times[1:] {
		if t.Before(min) {
			min = t
		}
	}

	return min
}

// maxHour returns the maximum hour from the given time.Time arguments
func maxHour(times ...time.Time) time.Time {
	if len(times) == 0 {
		// No arguments provided, return a zero time
		return time.Time{}
	}

	max := times[0]
	for _, t := range times[1:] {
		if t.After(max) {
			max = t
		}
	}

	return max
}

func convertToApiPOI(poi *POI) ApiPOI {
	apiPOI := ApiPOI{
		Lon:          poi.Lon,
		Lat:          poi.Lat,
		Name:         poi.Name,
		OpenHour:     convertTimeMapToString(poi.OpenHour),
		CloseHour:    convertTimeMapToString(poi.CloseHour),
		Satisfaction: poi.Satisfaction,
	}
	return apiPOI
}

func convertTimeMapToString(timeMap map[string]time.Time) map[string]string {
	stringMap := make(map[string]string)
	for key, value := range timeMap {
		stringMap[key] = value.Format("15:04")
	}
	return stringMap
}

func convertToApiItinerary(itinerary *Itinerary) ApiItinerary {
	apiItinerary := ApiItinerary{
		DayBeginHour: itinerary.DayBeginHour.Format("15:04"),
		DayEndHour:   itinerary.DayEndHour.Format("15:04"),
	}

	for _, day := range itinerary.Days {
		apiDay := ApiDay{
			DayNumber: day.DayNumber,
			DayName:   day.DayName,
		}

		for _, visit := range day.Visits {
			apiVisit := ApiVisit{
				Poi:           convertToApiPOI(visit.Poi),
				StartVisit:    visit.StartVisit.Format("15:04"),
				EndVisit:      visit.EndVisit.Format("15:04"),
				VisitDuration: visit.VisitDuration,
			}
			apiDay.Visits = append(apiDay.Visits, apiVisit)
		}

		apiItinerary.Days = append(apiItinerary.Days, apiDay)
	}

	return apiItinerary
}
