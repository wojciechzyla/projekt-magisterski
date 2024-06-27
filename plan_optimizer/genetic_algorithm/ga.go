package genetic_algorithm

import (
	"encoding/csv"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type solution struct {
	itinerary      Itinerary
	age            int
	objectiveValue float64
}

type GeneticAlgorithm struct {
	population             []solution
	poiList                []*POI
	constraints            Constraint
	dayBeginHour           time.Time
	dayEndHour             time.Time
	daysList               []string
	poiMultiplier          float64
	penaltyMultiplier      float64
	satisfactionMultiplier float64
}

const MUTATION_PROBABILITY = 0.2
const DAY_MUTATION_PROBABILITY = 0.6

func saveValuesToFile(bestValues []float64) {
	fileName := "results.csv"

	// Create the CSV file or open it in append mode
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Check if the file is new and write the header if necessary
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	if fileInfo.Size() == 0 {
		// Write the header if the file is empty
		header := []string{"iteration", "best_value"}
		if err := writer.Write(header); err != nil {
			panic(err)
		}
	}

	// Append records for each iteration
	for iteration, bestValue := range bestValues {
		record := []string{
			strconv.Itoa(iteration + 1), // Iteration numbers typically start from 1
			strconv.FormatFloat(bestValue, 'f', 10, 64),
		}
		if err := writer.Write(record); err != nil {
			panic(err)
		}
	}
}

func CreateGeneticAlgorithm(dayBeginHour, dayEndHour time.Time, daysList []string, poiMultiplier float64,
	penaltyMultiplier float64, satisfactionMultiplier float64) (ga *GeneticAlgorithm) {
	visitsWithinDayLimits := &VisitsWithinDayLimits{}

	timeDifferenceBetweenPoints := &TimeDifferenceBetweenPoints{}
	timeDifferenceBetweenPoints.setNext(visitsWithinDayLimits)

	poiOpenedDuringVisit := &PoiOpenedDuringVisit{}
	poiOpenedDuringVisit.setNext(timeDifferenceBetweenPoints)

	minimumTimeInPoi := &MinimumTimeInPoi{}
	minimumTimeInPoi.setNext(poiOpenedDuringVisit)

	originalPoi := &OriginalPoi{}
	originalPoi.setNext(minimumTimeInPoi)

	ga = &GeneticAlgorithm{
		constraints:            originalPoi,
		dayBeginHour:           dayBeginHour,
		dayEndHour:             dayEndHour,
		daysList:               daysList,
		poiMultiplier:          poiMultiplier,
		penaltyMultiplier:      penaltyMultiplier,
		satisfactionMultiplier: satisfactionMultiplier,
	}
	return ga
}

func (ga *GeneticAlgorithm) AddPoi(p *POI) {
	ga.poiList = append(ga.poiList, p)
}

func (ga *GeneticAlgorithm) selectParentsPairs(numberOfPairs int) [][]solution {
	totalFitness := 0.0
	for _, sol := range ga.population {
		totalFitness += sol.objectiveValue
	}
	result := make([][]solution, len(ga.population))
	for i := 0; i < numberOfPairs; i++ {
		result[i] = make([]solution, 0)
		for j := 0; j < 2; j++ {
			// Generate a random value between 0 and the total fitness
			randomValue := rand.Float64() * totalFitness

			// Iterate through the solutions and accumulate fitness until the random value is reached
			accumulatedFitness := 0.0
			for _, sol := range ga.population {
				accumulatedFitness += sol.objectiveValue
				if accumulatedFitness >= randomValue {
					result[i] = append(result[i], sol)
					break
				}
			}
		}
	}
	return result
}

func (ga *GeneticAlgorithm) sortPopulation() {
	sort.Slice(ga.population, func(i, j int) bool {
		return ga.population[i].objectiveValue > ga.population[j].objectiveValue
	})
}

func (ga *GeneticAlgorithm) Run(initialPopulationSize int, iterations int, solutionTTL int) ApiItinerary {
	ga.createInitialPopulation(initialPopulationSize)
	rand.Seed(time.Now().UnixNano())
	var solutionsToDelete []int
	var numberOfParents int
	bestObjectiveValue := -1000000.0

	var bestItinerary Itinerary
	//bestValuesSlice := make([]float64, iterations)

	for i := 0; i < iterations; i++ {
		if len(ga.daysList) > 1 {
			numberOfParents = rand.Intn((initialPopulationSize/5)-(initialPopulationSize/10)) + (initialPopulationSize / 10) + 1
			parents := ga.selectParentsPairs(numberOfParents)

			solutionChan := make(chan solution, len(parents))
			var wg sync.WaitGroup

			for _, pair := range parents {
				if len(pair) == 0 {
					continue
				}
				wg.Add(1)
				go func(pair []solution) {
					defer wg.Done()
					newItinerary1, newItinerary2 := CrossoverMultipleDays(&pair[0].itinerary, &pair[1].itinerary, ga.poiList)
					newSolution1 := solution{
						itinerary:      newItinerary1,
						age:            0,
						objectiveValue: 0.0,
					}
					newSolution2 := solution{
						itinerary:      newItinerary2,
						age:            0,
						objectiveValue: 0.0,
					}

					if rand.Float64() < MUTATION_PROBABILITY {
						substitutePOI(&newSolution1, ga.poiList, DAY_MUTATION_PROBABILITY)
					}

					if rand.Float64() < MUTATION_PROBABILITY {
						substitutePOI(&newSolution2, ga.poiList, DAY_MUTATION_PROBABILITY)
					}
					solutionChan <- newSolution1
					solutionChan <- newSolution2
				}(pair)
			}

			go func() {
				wg.Wait()
				close(solutionChan)
			}()

			for sol := range solutionChan {
				ga.population = append(ga.population, sol)
			}

			solutionsToDelete = make([]int, 0)
			for solId, sol := range ga.population {
				if sol.age > solutionTTL {
					solutionsToDelete = append(solutionsToDelete, solId)
				}
				ga.population[solId].age += 1
			}

			for _, idToDelete := range solutionsToDelete {
				// delete too old solutions
				if idToDelete < len(ga.population)-1 {
					ga.population = append(ga.population[:idToDelete], ga.population[idToDelete+1:]...)
				} else {
					ga.population = ga.population[:idToDelete]
				}
			}

			if len(ga.population) > initialPopulationSize {
				// delete the weakest solution if populations is too large
				ga.population = ga.population[:initialPopulationSize]
			}

		} else {
			var wg sync.WaitGroup
			for i, _ := range ga.population {
				wg.Add(1)
				go func(data *solution) {
					defer wg.Done()
					substitutePOI(data, ga.poiList, 0.8)
				}(&ga.population[i])
			}
			go func() {
				wg.Wait()
			}()
		}

		ga.assessPopulation()
		ga.sortPopulation()

		if i == 0 || ga.population[0].objectiveValue > bestObjectiveValue {
			bestObjectiveValue = ga.population[0].objectiveValue
			bestItinerary = ga.population[0].itinerary
		}
		//bestValuesSlice[i] = bestObjectiveValue
		// last := len(ga.population) - 1
		// fmt.Printf("Iteration %d\n Best: %f\n Worst: %f\n Best ever: %f\n Size: %d\n", i, ga.population[0].objectiveValue,
		// 	ga.population[last].objectiveValue, bestObjectiveValue, len(ga.population))
	}

	//saveValuesToFile(bestValuesSlice)
	return convertToApiItinerary(&bestItinerary)
}

func (ga *GeneticAlgorithm) assessPopulation() {
	for i, s := range ga.population {
		failedConstraints := ConstraintsCount{}
		ga.constraints.execute(&s.itinerary, &failedConstraints)

		objectiveFunction(&ga.population[i], failedConstraints.failedConstraints, ga.poiMultiplier, ga.penaltyMultiplier,
			ga.satisfactionMultiplier)
	}
}

//func (ga *GeneticAlgorithm) assessSingleSolution(s *solution) {
//	failedConstraints := ConstraintsCount{}
//	ga.constraints.execute(&s.itinerary, &failedConstraints)
//	objectiveFunction(s, failedConstraints.failedConstraints, ga.poiMultiplier, ga.penaltyMultiplier)
//}

func (ga *GeneticAlgorithm) createInitialPopulation(populationSize int) {
	rand.Seed(time.Now().UnixNano())

	solutionChan := make(chan solution, populationSize)
	var wg sync.WaitGroup

	ga.population = nil
	for i := 0; i < populationSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			itinerary := GenerateRandomItinerary(ga.poiList, ga.dayBeginHour, ga.dayEndHour, ga.daysList)
			newSolution := solution{
				itinerary:      itinerary,
				age:            0,
				objectiveValue: 0.0,
			}
			solutionChan <- newSolution
		}()
	}
	go func() {
		wg.Wait()
		close(solutionChan)
	}()

	ga.population = make([]solution, 0, populationSize)
	for sol := range solutionChan {
		ga.population = append(ga.population, sol)
	}
	ga.assessPopulation()
}
