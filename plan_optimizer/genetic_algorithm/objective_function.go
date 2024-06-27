package genetic_algorithm

func objectiveFunction(s *solution, failedConstraints int, poiMultiplier float64, penaltyMultiplier float64,
	satisfactionMultiplier float64) {
	var satisfaction float64
	var numberOfPoi = 0.0

	for _, day := range s.itinerary.Days {
		for _, visit := range day.Visits {
			numberOfPoi += 1.0
			satisfaction += float64(visit.VisitDuration) / (24.0 * 60.0) * visit.Poi.Satisfaction
		}
	}
	s.objectiveValue = satisfactionMultiplier*satisfaction + numberOfPoi*poiMultiplier - penaltyMultiplier*float64(failedConstraints)
}
