package sysstats

import (
	"runtime"
	"time"

	"github.com/yangchenxing/foochow/stats"
)

func AddSysStats(prefix string, interval time.Duration) {
	numCPU := stats.NewValue(prefix + "NumCPU")
	numGoroutine := stats.NewValue(prefix + "NumGoroutine")

	alloc := stats.NewValue(prefix + "Alloc")
	totalAlloc := stats.NewValue(prefix + "TotalAlloc")
	sys := stats.NewValue(prefix + "Sys")

	nextGC := stats.NewValue(prefix + "NextGC")
	lastGC := stats.NewValue(prefix + "LastGC")
	pauseTotalNs := stats.NewValue(prefix + "PauseTotalNs")
	lastPauseNs := stats.NewValue(prefix + "LastPauseNs")
	numGC := stats.NewValue(prefix + "NumGC")
	// gcCPUFraction := stats.NewValue(prefix + "GCCPUFraction")

	go func() {
		now := time.Now()
		time.Sleep(interval - now.Sub(now.Truncate(interval)))
		var memStats runtime.MemStats
		for {
			numCPU.Set(float64(runtime.NumCPU()))
			numGoroutine.Set(float64(runtime.NumGoroutine()))

			runtime.ReadMemStats(&memStats)
			alloc.Set(float64(memStats.Alloc))
			totalAlloc.Set(float64(memStats.TotalAlloc))
			sys.Set(float64(memStats.Sys))

			nextGC.Set(float64(memStats.NextGC))
			lastGC.Set(float64(memStats.LastGC))
			pauseTotalNs.Set(float64(memStats.PauseTotalNs))
			lastPauseNs.Set(float64(memStats.PauseNs[(memStats.NumGC+255)%256]))
			numGC.Set(float64(memStats.NumGC))
			// gcCPUFraction.Set(float64(memStats.GCCPUFraction))

			time.Sleep(interval)
		}
	}()
}
