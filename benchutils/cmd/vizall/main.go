package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/benchmark/parse"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type Result struct {
	Channels, Elements int
	Implementation     string
	NsPerOp            float64
}

func (r Result) NsPerOpPerElem() float64 {
	return r.NsPerOp / float64(r.Elements)
}

func ToValues(rs []Result) plotter.Values {
	out := make([]float64, len(rs))
	for i := range out {
		out[i] = rs[i].NsPerOpPerElem()
	}
	return plotter.Values(out)
}

type ResultGroup struct {
	Name    string
	Results []Result
}

func main() {
	set, err := parse.ParseSet(os.Stdin)
	if err != nil {
		log.Fatalf("failed parsing stdin: %v", err)
	}
	results := make(map[string][]Result)
	for key, value := range set {
		benchmark := strings.Split(key, "/")[1] // get the non-generic part of the name
		benchmark = strings.TrimSuffix(benchmark, "-4")
		parts := strings.Split(benchmark, ",")
		channels, _ := strconv.Atoi(strings.Split(parts[0], ":")[1])
		elements, _ := strconv.Atoi(strings.Split(parts[1], ":")[1])
		implementation := strings.Split(parts[2], ":")[1]
		nsOp := value[0].NsPerOp
		implName := implementation + fmt.Sprintf("%d", channels)
		results[implName] = append(results[implName], Result{Channels: channels, Elements: elements, Implementation: implementation, NsPerOp: nsOp})
	}
	sortedResults := make([]ResultGroup, 0, len(results))
	for setName, resultList := range results {
		sort.Slice(resultList, func(i, j int) bool {
			return resultList[i].Elements < resultList[j].Elements
		})
		sortedResults = append(sortedResults, ResultGroup{Name: setName, Results: resultList})
	}
	sort.Slice(sortedResults, func(i, j int) bool {
		return strings.Compare(sortedResults[i].Name, sortedResults[j].Name) < 1
	})
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Fan-In implementation time-per-element (lower is better)"
	p.Y.Label.Text = "ns/element"
	p.X.Label.Text = "elements sent"
	xlabels := []string{}
	for _, data := range sortedResults[0].Results {
		xlabels = append(xlabels, fmt.Sprintf("%d", data.Elements))
	}
	p.NominalX(xlabels...)
	barWidth := vg.Points(16)

	currentIndex := 0
	colorMap := map[string]color.RGBA{
		"concrete1":         color.RGBA{R: 200, G: 50, A: 255},
		"concrete10":        color.RGBA{R: 230, G: 100, A: 255},
		"concrete100":       color.RGBA{R: 255, G: 150, A: 255},
		"hybrid-closure1":   color.RGBA{G: 200, R: 150, A: 255},
		"hybrid-closure10":  color.RGBA{G: 230, A: 255},
		"hybrid-closure100": color.RGBA{G: 255, B: 150, A: 255},
		"hybrid-reflect1":   color.RGBA{B: 200, G: 150, A: 255},
		"hybrid-reflect10":  color.RGBA{B: 230, R: 100, G: 100, A: 255},
		"hybrid-reflect100": color.RGBA{B: 255, R: 150, A: 255},
	}
	for _, data := range sortedResults {
		values := ToValues(data.Results)

		bars, err := plotter.NewBarChart(values, barWidth)
		if err != nil {
			panic(err)
		}
		bars.Color = colorMap[data.Name]
		bars.Offset = barWidth * vg.Length(float64(currentIndex))
		p.Add(bars)
		p.Legend.Add(data.Name, bars)
		currentIndex++
	}
	p.Add(plotter.NewGlyphBoxes())
	err = p.Save(1000, 500, "testplot.png")
	if err != nil {
		panic(err)
	}
}
