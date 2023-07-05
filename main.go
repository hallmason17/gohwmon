package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	psCpu "github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	BYTES_IN_GB = 1024 * 1024 * 1024
	grid        = ui.NewGrid()
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to init termui: %v", err)
	}
	defer ui.Close()
	rate := 0.5
	interval := time.Second * time.Duration(1/rate)
	updateInt := time.NewTicker(interval).C
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	update(interval)
	uiEvents := ui.PollEvents()
	for {
		select {
		case <-sigTerm:
			return
		case <-updateInt:
			go update(interval)
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		}
	}
}

func update(updateInterval time.Duration) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err)
	}
	cpuPcnt, err := psCpu.Percent(updateInterval, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("\033[H\033[2J")
	fmt.Println("Memory:")
	fmt.Printf(
		"Total: %.2fGB, Free: %.2vGB, Used: %.2fGB UsedPercent:%.2f%%\n",
		float64(memory.Total)/float64(BYTES_IN_GB),
		float64(memory.Free)/float64(BYTES_IN_GB),
		float64(memory.Total)/float64(BYTES_IN_GB)*(memory.UsedPercent/100),
		memory.UsedPercent,
	)
	fmt.Println("CPU:")
	for i := 0; i < len(cpuPcnt); i++ {
		fmt.Printf("core%v %3.0f%%\n", i, cpuPcnt[i])
	}

	cycle_count, err := os.ReadFile("/sys/class/power_supply/BAT0/cycle_count")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Battery Cycle Count: %s", string(cycle_count))

	charging, err := os.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		fmt.Println(err)
	}
	if string(charging) != "Discharging\n" {
		fmt.Println("Battery is not discharging, cannot measure power consumption.")
		return
	}
	energy_now, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_now")
	if err != nil {
		fmt.Println(err)
	}
	power_now, err := os.ReadFile("/sys/class/power_supply/BAT0/power_now")
	if err != nil {
		fmt.Println(err)
	}
	capacity, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_full")
	energy_now1 := string(energy_now)
	energy_now1 = energy_now1[:len(energy_now1)-2]
	energy, err := strconv.Atoi(string(energy_now1))
	if err != nil {
		fmt.Println(err)
	}

	power_now1 := string(power_now)
	power_now1 = power_now1[:len(power_now1)-2]
	power, err := strconv.Atoi(string(power_now1))
	if err != nil {
		fmt.Println(err)
	}
	capacity1 := string(capacity)
	capacity1 = capacity1[:len(capacity1)-2]
	capacity2, err := strconv.Atoi(string(capacity1))
	if err != nil {
		fmt.Println(err)
	}

	bat_time_left := float64(energy) / float64(power) * 3600

	fmt.Printf("Battery Level: %.2f%%\n", float64(energy)/float64(capacity2)*100)
	fmt.Printf("Battery Time Left: %.2f hours\n", bat_time_left/60/60)
	fmt.Printf("Current power consumption: %3.2fW", float64(power)/100000)
}

func SetupGrid(data []float64) {
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	data1 := widgets.NewSparkline()
	data1.Data = data
	data1.LineColor = ui.ColorRed
	slg := widgets.NewSparklineGroup(data1)

	var cpuRow ui.GridItem

	cpuRow = ui.NewRow(1.0/2,
		ui.NewCol(1.0/2, slg),
		ui.NewCol(1.0/2, slg),
	)

	grid.Set(
		cpuRow,
	)
	ui.Render(grid)
}
