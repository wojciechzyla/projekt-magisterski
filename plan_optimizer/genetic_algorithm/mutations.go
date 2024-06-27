package genetic_algorithm

import (
	"math/rand"
	"time"
)

func substitutePOI(sol *solution, allPois []*POI, mutationProbability float64) {
	rand.Seed(time.Now().UnixNano())

	// Filter out POIs already used in the solution
	unusedPois := filterUnusedPois(sol, allPois)

	// If the solution has only one day, randomly select one visit and try to exchange it
	if len(sol.itinerary.Days) == 1 {
		day := &sol.itinerary.Days[0]
		if len(day.Visits) > 0 {
			visitId := rand.Intn(len(day.Visits))
			trySubstituteVisit(day, visitId, unusedPois, sol.itinerary.DayBeginHour, sol.itinerary.DayEndHour)
		}
	} else {
		// If there are more than one day, apply the mutation with some probability for each day
		for i, day := range sol.itinerary.Days {
			if len(day.Visits) > 0 && rand.Float64() < mutationProbability {
				visitId := rand.Intn(len(day.Visits))
				unusedPois = trySubstituteVisit(&sol.itinerary.Days[i], visitId, unusedPois, sol.itinerary.DayBeginHour, sol.itinerary.DayEndHour)
			}
		}
	}
}

func filterUnusedPois(sol *solution, allPois []*POI) []*POI {
	// Extract all POIs used in the solution
	usedPois := make(map[*POI]bool)
	for _, day := range sol.itinerary.Days {
		for _, visit := range day.Visits {
			usedPois[visit.Poi] = true
		}
	}

	// Filter out used POIs from the allPois list
	unusedPois := make([]*POI, 0)
	for i, poi := range allPois {
		if _, used := usedPois[poi]; !used {
			unusedPois = append(unusedPois, allPois[i])
		}
	}
	return unusedPois
}

func trySubstituteVisit(day *Day, visitId int, unusedPois []*POI, dayBeginHour, dayEndHour time.Time) []*POI {
	visit := &day.Visits[visitId]

	// Shuffle the unusedPois in random order
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(unusedPois), func(i, j int) {
		unusedPois[i], unusedPois[j] = unusedPois[j], unusedPois[i]
	})
	for i, newPoi := range unusedPois {
		result, visitStart, visitEnd := doesNewPoiFit(newPoi, day.Visits, visitId, dayBeginHour, dayEndHour, day.DayName)
		if result {
			// Substitute the visit with the new POI
			visit.Poi = newPoi
			visit.StartVisit = visitStart
			visit.EndVisit = visitEnd

			// Delete the updated POI from unusedPois
			unusedPois[i] = unusedPois[len(unusedPois)-1]
			unusedPois = unusedPois[:len(unusedPois)-1]

			return unusedPois
		}
	}

	return unusedPois
}
