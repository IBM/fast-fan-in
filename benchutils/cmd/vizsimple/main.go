package main

import (
	"flag"
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
	var (
		elementsFilter                     int
		filterByChannels, filterByElements bool
		channelsFilter                     int
		outputFileName                     string
	)
	flag.IntVar(&channelsFilter, "channels", -1, "Simplify the graph using only benchmarks with which number of channels")
	flag.IntVar(&elementsFilter, "elements", -1, "Simplify the graph using only benchmarks with which number of elements")
	flag.StringVar(&outputFileName, "output", "simple-viz.png", "The name of the output file name to use")
	flag.Parse()
	if channelsFilter != -1 {
		filterByChannels = true
	}
	if elementsFilter != -1 {
		filterByElements = true
	}
	if filterByChannels == filterByElements && !filterByChannels {
		log.Fatalf("must filter by channels, elements, or both")
	}
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
		if filterByChannels && channels != channelsFilter {
			continue
		}
		elements, _ := strconv.Atoi(strings.Split(parts[1], ":")[1])
		if filterByElements && elements != elementsFilter {
			continue
		}
		implementation := strings.Split(parts[2], ":")[1]
		nsOp := value[0].NsPerOp
		implName := implementation
		results[implName] = append(results[implName], Result{Channels: channels, Elements: elements, Implementation: implementation, NsPerOp: nsOp})
	}
	sortedResults := make([]ResultGroup, 0, len(results))
	for setName, resultList := range results {
		sort.Slice(resultList, func(i, j int) bool {
			if filterByChannels {
				return resultList[i].Elements < resultList[j].Elements
			} else {
				return resultList[i].Channels < resultList[j].Channels
			}
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
	p.Title.Text = fmt.Sprintf("Fan-In implementation time-per-element sent")
	if filterByChannels && filterByElements {
		p.Title.Text += fmt.Sprintf(" with %d elements sent over %d channels", elementsFilter, channelsFilter)
	} else if filterByChannels {
		p.Title.Text += fmt.Sprintf(" with %d managed channels", channelsFilter)
	} else if filterByElements {
		p.Title.Text += fmt.Sprintf(" with %d elements sent", elementsFilter)
	}
	p.Title.Text += " (lower is better)"
	p.Y.Label.Text = "ns/element"
	if filterByChannels && filterByElements {
		p.X.Label.Text = "implementation"
	} else if filterByChannels {
		p.X.Label.Text = "elements sent"
	} else if filterByElements {
		p.X.Label.Text = "channels managed"
	}
	xlabels := []string{}
	for _, data := range sortedResults[0].Results {
		if filterByChannels && filterByElements {
			xlabels = append(xlabels, data.Implementation)
		} else if filterByChannels {
			xlabels = append(xlabels, fmt.Sprintf("%d", data.Elements))
		} else if filterByElements {
			xlabels = append(xlabels, fmt.Sprintf("%d", data.Channels))
		}
	}
	p.NominalX(xlabels...)
	barWidth := vg.Points(16)

	currentIndex := 0
	colorMap := map[string]color.RGBA{
		"concrete":       color.RGBA{R: 255, G: 150, A: 255},
		"hybrid-closure": color.RGBA{G: 255, B: 150, A: 255},
		"hybrid-reflect": color.RGBA{B: 255, R: 150, A: 255},
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
	err = p.Save(500, 500, outputFileName)
	if err != nil {
		panic(err)
	}
}
