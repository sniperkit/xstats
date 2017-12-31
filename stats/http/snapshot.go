package httpstats

import (
	"fmt"
	"sync"
	"time"

	"github.com/sniperkit/xstats/pkg"
	// "github.com/jamiealquiza/tachymeter"
)

var Stats statistics

func init() {
	Stats = statistics{
		lock: sync.RWMutex{},
		numberOfRequestsByStatusCode:  make(map[int]int),
		numberOfRequestsByContentType: make(map[string]int),
	}
}

func New() *statistics {
	return &statistics{
		lock: sync.RWMutex{},
		numberOfRequestsByStatusCode:  make(map[int]int),
		numberOfRequestsByContentType: make(map[string]int),
	}
}

func updateStatistics(s time.Time, e time.Time, m *metrics) {
	go Stats.Add(s, e, *m)
}

type Snapshot struct {
	// time
	timestamp           time.Time
	timeSinceStart      time.Duration
	averageResponseTime time.Duration

	// counters
	numberOfWorkers              int
	totalNumberOfRequests        int
	numberOfSuccessfulRequests   int
	numberOfUnsuccessfulRequests int
	numberOfRequestsPerSecond    float64

	numberOfRequestsByStatusCode  map[int]int
	numberOfRequestsByContentType map[string]int

	// size
	totalSizeInBytes   int
	averageSizeInBytes int
}

func (snapshot Snapshot) Timestamp() time.Time {
	return snapshot.timestamp
}

func (snapshot Snapshot) NumberOfWorkers() int {
	return snapshot.numberOfWorkers
}

func (snapshot Snapshot) NumberOfErrors() int {
	return snapshot.numberOfUnsuccessfulRequests
}

func (snapshot Snapshot) TotalNumberOfRequests() int {
	return snapshot.totalNumberOfRequests
}

func (snapshot Snapshot) TotalSizeInBytes() int {
	return snapshot.totalSizeInBytes
}

func (snapshot Snapshot) AverageSizeInBytes() int {
	return snapshot.averageSizeInBytes
}

func (snapshot Snapshot) AverageResponseTime() time.Duration {
	return snapshot.averageResponseTime
}

func (snapshot Snapshot) RequestsPerSecond() float64 {
	return snapshot.numberOfRequestsPerSecond
}

type statistics struct {
	lock sync.RWMutex

	rawResults  []metrics
	snapShots   []Snapshot
	logMessages []string

	startTime time.Time
	endTime   time.Time

	totalResponseTime time.Duration

	numberOfWorkers               int
	numberOfRequests              int
	numberOfSuccessfulRequests    int
	numberOfUnsuccessfulRequests  int
	numberOfRequestsByStatusCode  map[int]int
	numberOfRequestsByContentType map[string]int

	totalSizeInBytes int
}

func (statistics *statistics) Update(s time.Time, e time.Time, m metrics) {
	statistics.lock.Lock() // // update the raw results
	defer statistics.lock.Unlock()

	statistics.Add(s, e, m)
}

func (statistics *statistics) Add(s time.Time, e time.Time, m metrics) Snapshot {
	statistics.lock.Lock() // // update the raw results
	defer statistics.lock.Unlock()
	statistics.rawResults = append(statistics.rawResults, m)

	// initialize start and end time
	if statistics.numberOfRequests == 0 {
		statistics.startTime = m.StartTime()
		statistics.endTime = m.EndTime()
	}

	// start time
	if m.StartTime().Before(statistics.startTime) {
		statistics.startTime = m.StartTime()
	}

	// end time
	if m.EndTime().After(statistics.endTime) {
		statistics.endTime = m.EndTime()
	}

	// update the total number of requests
	statistics.numberOfRequests = len(statistics.rawResults)

	// is successful
	if m.StatusCode() > 199 && m.StatusCode() < 400 {
		statistics.numberOfSuccessfulRequests += 1
	} else {
		statistics.numberOfUnsuccessfulRequests += 1
	}

	// number of workers
	statistics.numberOfWorkers = 1 // m.NumberOfWorkers()

	// number of requests by status code
	statistics.numberOfRequestsByStatusCode[m.StatusCode()] += 1

	// number of requests by content type
	statistics.numberOfRequestsByContentType[m.ContentType()] += 1

	// update the total duration
	responseTime := m.EndTime().Sub(m.StartTime())
	statistics.totalResponseTime += responseTime

	// size
	statistics.totalSizeInBytes += m.Size()
	averageSizeInBytes := statistics.totalSizeInBytes / statistics.numberOfRequests

	// average response time
	averageResponseTime := time.Duration(statistics.totalResponseTime.Nanoseconds() / int64(statistics.numberOfRequests))

	// number of requests per second
	requestsPerSecond := float64(statistics.numberOfRequests) / statistics.endTime.Sub(statistics.startTime).Seconds()

	// log messages
	statistics.logMessages = append(statistics.logMessages, m.ConsoleLog())

	// create a snapshot
	snapShot := Snapshot{

		// times
		timestamp:           m.EndTime(),
		averageResponseTime: averageResponseTime,

		// counters
		numberOfWorkers:               statistics.numberOfWorkers,
		totalNumberOfRequests:         statistics.numberOfRequests,
		numberOfSuccessfulRequests:    statistics.numberOfSuccessfulRequests,
		numberOfUnsuccessfulRequests:  statistics.numberOfUnsuccessfulRequests,
		numberOfRequestsPerSecond:     requestsPerSecond,
		numberOfRequestsByStatusCode:  statistics.numberOfRequestsByStatusCode,
		numberOfRequestsByContentType: statistics.numberOfRequestsByContentType,

		// size
		totalSizeInBytes:   statistics.totalSizeInBytes,
		averageSizeInBytes: averageSizeInBytes,
	}

	statistics.snapShots = append(statistics.snapShots, snapShot)

	return snapShot
}

func (statistics *statistics) LastSnapshot() Snapshot {
	statistics.lock.RLock()
	defer statistics.lock.RUnlock()

	lastSnapshotIndex := len(statistics.snapShots) - 1
	if lastSnapshotIndex < 0 {
		return Snapshot{}
	}

	return statistics.snapShots[lastSnapshotIndex]
}

func (statistics *statistics) LastLogMessages(count int) []string {
	statistics.lock.RLock()
	defer statistics.lock.RUnlock()

	messages, err := GetLatestLogMessages(statistics.logMessages, count)
	if err != nil {
		panic(err)
	}
	return messages
}

func GetLatestLogMessages(messages []string, count int) ([]string, error) {
	if count < 0 {
		return nil, fmt.Errorf("The count cannot be negative")
	}

	numberOfMessges := len(messages)
	if count == numberOfMessges {
		return messages, nil
	}
	if count < numberOfMessges {
		return messages[numberOfMessges-count:], nil
	}
	if count > numberOfMessges {
		fillLines := make([]string, count-numberOfMessges)
		return append(fillLines, messages...), nil
	}
	stats.Log.Entry.Fatal("model.GetLatestLogMessages(), Unreachable")
	panic("Unreachable")
}
