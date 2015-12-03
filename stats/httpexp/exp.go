package statshttp

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yangchenxing/foochow/stats"

	"bitbucket.org/jesgoo/avalon/encoding"
)

var (
	path       = "/stats"
	duration   = time.Minute
	data       map[string]float64
	updateTime time.Time
	lock       sync.Mutex
)

func SetupHandler(p string, d time.Duration) {
	path = p
	duration = d
	http.HandleFunc(p, handleExport)
}

func handleExport(response http.ResponseWriter, request *http.Request) {
	updateData()
	var prefix string
	if request.URL.Path != path {
		prefix = request.URL.Path[len(path)+1:]
	}
	result := make(map[string]float64)
	for key, value := range data {
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	response.Header().Set("Content-Type", "text/json; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(encoding.Jsonify(result)))
}

func updateData() {
	lock.Lock()
	defer lock.Unlock()
	if time.Now().Sub(updateTime) > duration/2 {
		data = stats.GetAllStatsItemValues(duration)
		updateTime = time.Now()
	}
}
