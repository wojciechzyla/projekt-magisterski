package genetic_algorithm

import (
	"math/rand"
	"time"
)

type PoiToChangeTuple struct {
	DayId   int
	VisitId int
	Poi     *POI
}

func createChildMultipleDays(source1, source2 *Itinerary, divisionIndex int) Itinerary {
	child := Itinerary{
		Days:         make([]Day, len(source1.Days)),
		DayBeginHour: source1.DayBeginHour,
		DayEndHour:   source1.DayEndHour,
	}

	for i := 0; i < divisionIndex; i++ {
		child.Days[i] = copyDay(source1.Days[i])
	}

	for i := divisionIndex; i < len(source1.Days); i++ {
		child.Days[i] = copyDay(source2.Days[i])
	}

	return child
}

func copyDay(original Day) Day {
	dayCopy := Day{
		Visits:    make([]Visit, len(original.Visits)),
		DayNumber: original.DayNumber,
		DayName:   original.DayName,
	}

	for j := 0; j < len(original.Visits); j++ {
		dayCopy.Visits[j] = copyVisit(original.Visits[j])
	}

	return dayCopy
}

func copyVisit(original Visit) Visit {
	return Visit{
		Poi:           original.Poi,
		StartVisit:    original.StartVisit,
		EndVisit:      original.EndVisit,
		VisitDuration: original.VisitDuration,
	}
}

func createChildren(itinerary1, itinerary2 *Itinerary) (Itinerary, Itinerary) {
	divisionIndex := rand.Intn(len(itinerary1.Days)-1) + 1
	child1 := createChildMultipleDays(itinerary1, itinerary2, divisionIndex)
	child2 := createChildMultipleDays(itinerary2, itinerary1, divisionIndex)
	return child1, child2
}

func CrossoverMultipleDays(itinerary1, itinerary2 *Itinerary, allPoiList []*POI) (Itinerary, Itinerary) {

	child1, child2 := createChildren(itinerary1, itinerary2)

	var newSolutions = []Itinerary{child1, child2}
	var newPoiToChange PoiToChangeTuple
	var updatedPoi *POI

	for _, itinerary := range newSolutions {
		usedPoiList := make([]*POI, 0)
		poiToChange := make([]PoiToChangeTuple, 0)

		for dayId, day := range itinerary.Days {
			for visitId, visit := range day.Visits {
				if containsPoi(usedPoiList, visit.Poi) {
					newPoiToChange = PoiToChangeTuple{DayId: dayId, VisitId: visitId, Poi: visit.Poi}
					poiToChange = append(poiToChange, newPoiToChange)
				} else {
					usedPoiList = append(usedPoiList, visit.Poi)
				}
			}
		}

		for changeId, change := range poiToChange {
			updatedPoi = nil
			newVisitStart := time.Time{}
			newVisitEnd := time.Time{}
			poiFit := false
			availablePoi := make([]*POI, 0)
			for _, poi := range allPoiList {
				if !containsPoi(usedPoiList, poi) {
					availablePoi = append(availablePoi, poi)
				}
			}

			dayId := change.DayId
			visitId := change.VisitId
			day := itinerary.Days[dayId]

			for len(availablePoi) > 0 {
				newPoi, newPoiIndex := drawPoi(availablePoi)
				poiFit, newVisitStart, newVisitEnd = doesNewPoiFit(newPoi, day.Visits, visitId, itinerary.DayBeginHour, itinerary.DayEndHour, day.DayName)
				if poiFit {
					usedPoiList = append(usedPoiList, newPoi)
					updatedPoi = newPoi
					break
				} else {
					// delete newPoi from availablePoi list
					availablePoi[newPoiIndex] = availablePoi[len(availablePoi)-1]
					availablePoi = availablePoi[:len(availablePoi)-1]
				}
			}

			if updatedPoi == nil {
				currentVisit := &itinerary.Days[dayId].Visits[visitId]
				if visitId > 0 {
					prevVisit := &itinerary.Days[dayId].Visits[visitId-1]
					itinerary.Days[dayId].Visits[visitId-1].EndVisit = minHour(addMinutes(prevVisit.EndVisit,
						currentVisit.VisitDuration/2), prevVisit.Poi.CloseHour[day.DayName])
				}
				if visitId < len(itinerary.Days[dayId].Visits)-1 {
					nextVisit := &itinerary.Days[dayId].Visits[visitId+1]
					itinerary.Days[dayId].Visits[visitId+1].StartVisit = maxHour(subtractMinutes(nextVisit.StartVisit,
						currentVisit.VisitDuration/2), nextVisit.Poi.OpenHour[day.DayName])
				}
				itinerary.Days[dayId].Visits = append(itinerary.Days[dayId].Visits[:visitId], itinerary.Days[dayId].Visits[visitId+1:]...)
				for ci := changeId + 1; ci < len(poiToChange); ci++ {
					if poiToChange[ci].DayId == dayId && poiToChange[ci].VisitId > visitId {
						poiToChange[ci].VisitId -= 1
					}
				}
			} else {
				itinerary.Days[dayId].Visits[visitId].Poi = updatedPoi
				itinerary.Days[dayId].Visits[visitId].StartVisit = newVisitStart
				itinerary.Days[dayId].Visits[visitId].EndVisit = newVisitEnd
			}
		}
	}

	return child1, child2
}
