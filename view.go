package main

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type view struct {
	Grid *ui.Grid

	InfoBar     *widgets.Paragraph
	CpuChart    *widgets.BarChart
	MemChart    *widgets.BarChart
	BatteryList *widgets.Paragraph
}

type ChartData struct {
	DataLabels []string
	Data       []float64
}

type Stats struct {
	CpuChart ChartData
	MemChart ChartData
}

func createBarChart() *widgets.BarChart {
	chart := widgets.NewBarChart()
	chart.Border = true
	return chart
}

func (cd ChartData) UpdateBarChart(uiChart *widgets.BarChart) {
	uiChart.Data = cd.Data
	numBars := len(cd.Data)
	uiChart.BarColors = make([]ui.Color, numBars)
	uiChart.LabelStyles = make([]ui.Style, numBars)
	uiChart.NumStyles = make([]ui.Style, numBars)
	uiChart.Labels = make([]string, numBars)
	for i := 0; i < numBars; i++ {
		uiChart.BarColors[i] = ui.ColorWhite
		uiChart.LabelStyles[i] = ui.Style{Fg: ui.ColorWhite, Bg: ui.ColorClear}
		uiChart.NumStyles[i] = ui.Style{Fg: ui.ColorBlack}
		uiChart.Labels[i] = fmt.Sprintf("%3s", cd.DataLabels[i])
	}
}

func (v *view) UpdateStats(charts Stats) {
	charts.CpuChart.UpdateBarChart(v.CpuChart)
	charts.MemChart.UpdateBarChart(v.MemChart)

	v.Render()
}

func (v *view) SetLayout() {
	v.Grid = ui.NewGrid()
	v.ResetSize()
	v.Grid.Set(
		ui.NewRow(1.0/12,
			ui.NewCol(1.0, v.InfoBar),
		),
		ui.NewRow(3.0/12,
			ui.NewCol(1.0, v.CpuChart),
		),
		ui.NewRow(5.0/12,
			ui.NewCol(4.0, v.MemChart),
			ui.NewCol(8.0, v.BatteryList),
		),
	)
}

func (v *view) ResetSize() {
	termWidth, termHeight := ui.TerminalDimensions()
	if termWidth > 20 {
		v.Grid.SetRect(0, 0, termWidth, termHeight)
	}
}

func (v *view) Render() {
	ui.Render(v.Grid)
}

func NewView() *view {
	view := view{}

	view.InfoBar = widgets.NewParagraph()
	view.InfoBar.Border = true
	view.InfoBar.Text = ""
	view.InfoBar.Title = "Go HW Monitor"

	view.CpuChart = createBarChart()
	view.CpuChart.Title = "CPU"

	view.MemChart = createBarChart()
	view.MemChart.Title = "MEM"

	view.BatteryList = widgets.NewParagraph()
	view.BatteryList.Border = true
	view.BatteryList.Text = ""
	view.BatteryList.Title = "Battery"

	return &view
}
