package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/subscription"
)

const pageTemplate = `
<html>
	<head>
		<title>Feed generator status</title>
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<div style="max-width: 800px; padding: 1rem;">
			<canvas id="evtPerSecondChart"></canvas>
		</div>
		<div style="max-width: 800px; padding: 1rem;">
			<canvas id="lagTimeChart"></canvas>
		</div>

		<script src="https://cdn.jsdelivr.net/npm/chart.js@^3"></script>
		<script src="https://cdn.jsdelivr.net/npm/luxon@^2"></script>
		<script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-luxon@^1"></script>
		
		<script>
		  const {unitLabel, eventsPerSecondData, lagTimeData, toCatchupData} = %s;
		
			new Chart(document.getElementById('evtPerSecondChart'), {
				type: 'line',
				data: {
					datasets: [{
						label: 'Events per second',
						data: eventsPerSecondData,
						borderColor: '#047857',
					}]
				},
				options: {
					scales: {
						y: {
							beginAtZero: true
						},
						x: {
							type: 'time'
						}
					}
				}
			});
			new Chart(document.getElementById('lagTimeChart'), {
				type: 'line',
				data: {
					datasets: [
						{
							label: 'Lag time',
							data: lagTimeData,
							borderColor: '#2563eb',
						},
						{
							label: 'To catch up',
							data: toCatchupData,
							borderColor: '#a21caf',
						}
					]
				},
				options: {
					scales: {
						y: {
							// beginAtZero: true
							// type: 'logarithmic'
							title: {
								display: true,
								text: unitLabel
							}
						},
						x: {
							type: 'time'
						}
					}
				}
			});
		</script>
	</body>
</html>
`

type StatusPage struct {
	processingStats *subscription.ProcessingStats
	database        *database.Database
}

type Point struct {
	X string  `json:"x"`
	Y float64 `json:"y"`
}

func (handler StatusPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var eventsPerSecondData []Point
	var lagTimeData []Point
	var toCatchupData []Point
	handler.processingStats.Iterate(func(run *subscription.ProcessingRun) {
		x := run.Timestamp.Format(time.RFC3339Nano)
		eventsPerSecondData = append(eventsPerSecondData, Point{
			X: x,
			Y: run.EventsPerSecond,
		})
		lagTimeData = append(lagTimeData, Point{
			X: x,
			Y: run.LagTime.Seconds(),
		})
		toCatchupData = append(toCatchupData, Point{
			X: x,
			Y: run.ToCatchUp.Seconds(),
		})
	})
	indexLagTime := lagTimeData[len(lagTimeData)/2].Y
	unitLabel := "Seconds"
	if indexLagTime > 7200 {
		unitLabel = "Hours"
		for i := 0; i < len(lagTimeData); i++ {
			lagTimeData[i].Y = lagTimeData[i].Y / 3600
			toCatchupData[i].Y = toCatchupData[i].Y / 3600
		}
	} else if indexLagTime > 120 {
		unitLabel = "Minutes"
		for i := 0; i < len(lagTimeData); i++ {
			lagTimeData[i].Y = lagTimeData[i].Y / 60
			toCatchupData[i].Y = toCatchupData[i].Y / 60
		}
	}
	data := make(map[string]any)
	data["unitLabel"] = unitLabel
	data["eventsPerSecondData"] = eventsPerSecondData
	data["lagTimeData"] = lagTimeData
	data["toCatchupData"] = toCatchupData
	jsonData, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(pageTemplate, string(jsonData))))
}
