package genetic_algorithm

import "fmt"

type ConstraintsCount struct {
	failedConstraints int
}

type Constraint interface {
	execute(itinerary *Itinerary, failed *ConstraintsCount)
	setNext(Constraint)
}

type VisitsWithinDayLimits struct {
	next Constraint
}

func (v *VisitsWithinDayLimits) execute(itinerary *Itinerary, failed *ConstraintsCount) {
	failedConstraint := false
	for _, day := range itinerary.Days {
		dayLen := len(day.Visits)
		if day.Visits[0].StartVisit.Before(itinerary.DayBeginHour) ||
			day.Visits[dayLen-1].EndVisit.After(itinerary.DayEndHour) {
			failedConstraint = true
			break
		}
	}
	if failedConstraint {
		failed.failedConstraints += 1
		fmt.Println("VisitsWithinDayLimits")
	}
	if v.next != nil {
		v.next.execute(itinerary, failed)
	}
}

func (v *VisitsWithinDayLimits) setNext(next Constraint) {
	v.next = next
}

type TimeDifferenceBetweenPoints struct {
	next Constraint
}

func (t *TimeDifferenceBetweenPoints) execute(itinerary *Itinerary, failed *ConstraintsCount) {
	failedConstraint := false
	for _, day := range itinerary.Days {
		for i := 1; i < len(day.Visits); i++ {
			transportTime := transport(day.Visits[i-1].Poi, day.Visits[i].Poi)
			minimumStartHour := addMinutes(day.Visits[i-1].EndVisit, transportTime)
			if day.Visits[i].StartVisit.Before(minimumStartHour) {
				failedConstraint = true
				break
			}
		}
	}
	if failedConstraint {
		failed.failedConstraints += 1
		fmt.Println("TimeDifferenceBetweenPoints")
	}
	if t.next != nil {
		t.next.execute(itinerary, failed)
	}
}

func (t *TimeDifferenceBetweenPoints) setNext(next Constraint) {
	t.next = next
}

type PoiOpenedDuringVisit struct {
	next Constraint
}

func (p *PoiOpenedDuringVisit) execute(itinerary *Itinerary, failed *ConstraintsCount) {
	failedConstraint := false
	for _, day := range itinerary.Days {
		for _, visit := range day.Visits {
			if !openedDuringHours(visit.Poi, visit.StartVisit, visit.EndVisit, day.DayName) {
				failedConstraint = true
				break
			}
		}
	}
	if failedConstraint {
		failed.failedConstraints += 1
		fmt.Println("PoiOpenedDuringVisit")
	}
	if p.next != nil {
		p.next.execute(itinerary, failed)
	}
}

func (p *PoiOpenedDuringVisit) setNext(next Constraint) {
	p.next = next
}

type MinimumTimeInPoi struct {
	next Constraint
}

func (m *MinimumTimeInPoi) execute(itinerary *Itinerary, failed *ConstraintsCount) {
	failedConstraint := false
	for _, day := range itinerary.Days {
		for _, visit := range day.Visits {
			if calculateDuration(visit.StartVisit, visit.EndVisit) < 60 {
				failedConstraint = true
				break
			}
		}
	}
	if failedConstraint {
		failed.failedConstraints += 1
		fmt.Println("MinimumTimeInPoi")
	}
	if m.next != nil {
		m.next.execute(itinerary, failed)
	}
}

func (m *MinimumTimeInPoi) setNext(next Constraint) {
	m.next = next
}

type OriginalPoi struct {
	next Constraint
}

func (o *OriginalPoi) execute(itinerary *Itinerary, failed *ConstraintsCount) {
	failedConstraint := false
	var usedPoi = make([]*POI, 0)
	for _, day := range itinerary.Days {
		for _, visit := range day.Visits {
			if containsPoi(usedPoi, visit.Poi) {
				failedConstraint = true
				break
			}
			usedPoi = append(usedPoi, visit.Poi)
		}
	}
	if failedConstraint {
		failed.failedConstraints += 1
		fmt.Println("OriginalPoi")
	}
	if o.next != nil {
		o.next.execute(itinerary, failed)
	}
}

func (o *OriginalPoi) setNext(next Constraint) {
	o.next = next
}
