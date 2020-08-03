package cmd

import (
	"sync"
	"time"
)

type Result struct {
	Duration time.Duration
	Err      error
}

//
type Stats struct {
	AvgTotal     float64 //Millisecond
	Fastest      float64 //Millisecond
	Slowest      float64 //Millisecond
	ReqPerSecond float64

	TotalTime   time.Duration
	TotalTimeMS float64 //Millisecond

	TotalReq   int64
	ErrorNum   int64
	SuccessNum int64
}

type RequestReport struct {
	objStart       time.Time
	aryResult      []*Result
	objResultMutex sync.Mutex
	objStats       Stats
}

func (ptrReport *RequestReport) Reset() {
	ptrReport.objStart = time.Now()
	ptrReport.aryResult = []*Result{}
}

func (ptrReport *RequestReport) Result(nDuration time.Duration, anyErr error) {
	var objResult Result
	objResult.Duration = nDuration
	objResult.Err = anyErr

	ptrReport.objResultMutex.Lock()
	ptrReport.aryResult = append(ptrReport.aryResult, &objResult)
	ptrReport.objResultMutex.Unlock()
}

func (ptrReport *RequestReport) Stats() Stats {

	ptrReport.objResultMutex.Lock()
	defer ptrReport.objResultMutex.Unlock()

	var objStats Stats
	var nFastest time.Duration = 0
	var nSlowest time.Duration = 0

	objStats.TotalTime = time.Since(ptrReport.objStart)
	objStats.TotalTimeMS = float64(objStats.TotalTime.Microseconds()) / float64(1000)
	for _, ptrResult := range ptrReport.aryResult {

		objStats.TotalReq++
		if ptrResult.Err != nil {
			objStats.ErrorNum++
		} else {
			objStats.SuccessNum++
		}

		if ptrResult.Duration > nSlowest {
			nSlowest = ptrResult.Duration
		}
		if ptrResult.Duration < nFastest {
			nFastest = ptrResult.Duration
		}
	}
	objStats.Fastest = float64(nFastest.Microseconds()) / float64(1000)
	objStats.Slowest = float64(nSlowest.Microseconds()) / float64(1000)
	objStats.AvgTotal = objStats.TotalTimeMS / float64(objStats.TotalReq)
	objStats.ReqPerSecond = float64(objStats.TotalReq*1000) / (objStats.TotalTimeMS)

	ptrReport.objStats = objStats
	return objStats
}
