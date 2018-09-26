package main

import (
	"net/http"
	"time"

	"github.com/shirou/gopsutil/net"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type Bandwidth struct {
	t           []float64
	recv, sent  []float64
	timeBetween time.Duration
}

func New() Bandwidth {
	return Bandwidth{
		t:           []float64{},
		recv:        []float64{},
		sent:        []float64{},
		timeBetween: 1000 * time.Millisecond,
	}
}

func (b *Bandwidth) Tick() []float64 {
	return b.t
}

func (b *Bandwidth) Recv() []float64 {
	return b.recv
}

func (b *Bandwidth) Sent() []float64 {
	return b.sent
}

// Start
func (b *Bandwidth) Start() {
	go b.start()
}

func (b *Bandwidth) start() {
	startTime := time.Now()
	previousStat, err := net.IOCounters(false)
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(b.timeBetween)
	for range ticker.C {
		currentStat, err := net.IOCounters(false)
		if err != nil {
			panic(err)
		}
		b.t = append(b.t, time.Since(startTime).Seconds())
		b.sent = append(b.sent,
			float64(currentStat[0].BytesSent-previousStat[0].BytesSent)/(1024*1024*b.timeBetween.Seconds()),
		)
		b.recv = append(b.recv,
			float64(currentStat[0].BytesRecv-previousStat[0].BytesRecv)/(1024*1024*b.timeBetween.Seconds()),
		)
		previousStat[0] = currentStat[0]
	}
}

func main() {
	b := New()
	b.Start()
	http.HandleFunc("/", func(res http.ResponseWriter, r *http.Request) {
		series := make([]chart.Series, 2)

		series[0] = chart.ContinuousSeries{
			Name:    "Sent",
			XValues: b.Tick(),
			YValues: b.Sent(),
			Style: chart.Style{
				Show:        true,                           //note; if we set ANY other properties, we must set this to true.
				StrokeColor: drawing.ColorRed,               // will supercede defaults
				FillColor:   drawing.ColorRed.WithAlpha(64), // will supercede defaults
			},
		}

		series[1] = chart.ContinuousSeries{
			Name:    "Received",
			XValues: b.Tick(),
			YValues: b.Recv(),
			Style: chart.Style{
				Show:        true,                             //note; if we set ANY other properties, we must set this to true.
				StrokeColor: drawing.ColorGreen,               // will supercede defaults
				FillColor:   drawing.ColorGreen.WithAlpha(64), // will supercede defaults
			},
		}

		graph := chart.Chart{
			XAxis: chart.XAxis{
				Name:      "Time (s)",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			YAxis: chart.YAxis{
				Name:      "Bandwidth (MB/s)",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				Range: &chart.ContinuousRange{
					Min: 0.0,
					Max: 15.0,
				},
			},
			Series: series,
		}
		graph.Elements = []chart.Renderable{
			chart.Legend(&graph),
		}
		res.Header().Set("Content-Type", "image/png")
		graph.Render(chart.PNG, res)
	})
	http.HandleFunc("/favico.ico", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte{})
	})
	http.ListenAndServe(":8080", nil)
}
