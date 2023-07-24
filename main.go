package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	ui "github.com/gizak/termui/v3"
	psCpu "github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	BYTES_IN_GB = 1024 * 1024 * 1024
	grid        = ui.NewGrid()
	stats       Stats
)

func main() {
	if err := ui.Init(); err != nil {
	}
	defer ui.Close()
	uiView := NewView()
	uiView.SetLayout()
	uiView.Render()
	statsChan := make(chan Stats)
	rate := 0.5
	interval := time.Second * time.Duration(1/rate)
	updateInt := time.NewTicker(interval).C
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	// update(interval)
	stats = UpdateStatsCharts(UpdateCPU(interval), UpdateMem())
	uiView.UpdateStats(stats)
	uiEvents := ui.PollEvents()
	for {
		select {
		case <-sigTerm:
			return
		case <-updateInt:
			uiView.UpdateStats(UpdateStatsCharts(UpdateCPU(interval), UpdateMem()))
			// go update(interval)
		case newStats := <-statsChan:
			currentStats := newStats
			uiView.UpdateStats(currentStats)
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		}
	}
}

func UpdateMem() []float64 {
	memory, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err)
	}
	return []float64{float64(memory.Used), float64(memory.Free)}
}

func UpdateCPU(updateInterval time.Duration) []float64 {
	cpuPcnt, err := psCpu.Percent(updateInterval, true)
	if err != nil {
		fmt.Println(err)
	}
	return cpuPcnt
}

func UpdateStatsCharts(
	cpuStats []float64,
	memStats []float64,
) (statMsg Stats) {
	statMsg.CpuChart = ChartData{}
	statMsg.MemChart = ChartData{}

	statMsg.CpuChart.DataLabels = make([]string, len(cpuStats))
	statMsg.CpuChart.Data = cpuStats

	statMsg.MemChart.DataLabels = make([]string, len(memStats))
	statMsg.MemChart.Data = memStats

	for i := 0; i < len(cpuStats); i++ {
		statMsg.CpuChart.DataLabels[i] = "C" + fmt.Sprint(i)
	}

	return statMsg
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
		"Used: %.2fGB\nFree: %.2vGB\nTotal: %.2fGB\nUsedPercent:%.2f%%\n",
		float64(memory.Total)/float64(BYTES_IN_GB)*(memory.UsedPercent/100),
		float64(memory.Free)/float64(BYTES_IN_GB),
		float64(memory.Total)/float64(BYTES_IN_GB),
		memory.UsedPercent,
	)
	fmt.Println("CPU:")
	for i := 0; i < len(cpuPcnt); i++ {
		if i < 10 {
			fmt.Printf("C0%v %3.0f%%\n", i, cpuPcnt[i])
		} else {
			fmt.Printf("C%v %3.0f%%\n", i, cpuPcnt[i])
		}
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
	energy_now, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_now")
	if err != nil {
		fmt.Println(err)
	}
	energy_now1 := string(energy_now)
	energy_now1 = energy_now1[:len(energy_now1)-2]
	energy, err := strconv.Atoi(string(energy_now1))
	if err != nil {
		fmt.Println(err)
	}

	capacity, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_full")
	capacity1 := string(capacity)
	capacity1 = capacity1[:len(capacity1)-2]
	capacity2, err := strconv.Atoi(string(capacity1))
	if err != nil {
		fmt.Println(err)
	}
	energy_full, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_full_design")
	if err != nil {
		fmt.Println(err)
	}
	energy_full1 := string(energy_full)
	energy_full1 = energy_full1[:len(energy_full1)-2]
	max_energy, err := strconv.Atoi(energy_full1)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Battery Level: %.2f%%\n", float64(energy)/float64(capacity2)*100)

	if string(charging) != "Not charging\n" {
		power_now, err := os.ReadFile("/sys/class/power_supply/BAT0/power_now")
		if err != nil {
			fmt.Println(err)
		}
		power_now1 := string(power_now)
		power_now1 = power_now1[:len(power_now1)-2]
		power, err := strconv.Atoi(string(power_now1))
		if err != nil {
			fmt.Println(err)
		}
		if string(charging) == "Charging\n" {
			minutesLeft := int32((float64(max_energy)-float64(energy))/float64(power)*60) % 60
			if minutesLeft < 10 {
				strMinutesLeft := fmt.Sprint(minutesLeft)
				strMinutesLeft = "0" + strMinutesLeft
				fmt.Printf(
					"Time until charged: %d:%s\n",
					int32((float64(max_energy)-float64(energy))/float64(power)),
					strMinutesLeft,
				)
			} else if minutesLeft >= 10 {
				fmt.Printf(
					"Time until charged: %d:%d\n",
					int32((float64(max_energy)-float64(energy))/float64(power)),
					int32((float64(max_energy)-float64(energy))/float64(power)*60)%60,
				)
			}
		}
		if string(charging) == "Discharging\n" {
			bat_time_left := float64(energy) / float64(power) * 3600

			if int32(bat_time_left/60)%60 < 10 {
				intMinutesLeft := int32(bat_time_left/60) % 60
				minutesLeft := "0" + fmt.Sprint(intMinutesLeft)
				fmt.Println(minutesLeft)
				fmt.Printf(
					"Battery Time Left: %d:%s\n",
					int32(bat_time_left/60/60),
					minutesLeft,
				)
			} else if int32(bat_time_left/60)%60 >= 10 {
				fmt.Printf(
					"Battery Time Left: %d:%d\n",
					int32(bat_time_left/60/60),
					int32(bat_time_left/60)%60,
				)
			}

			fmt.Printf("Current power consumption: %3.2fW", float64(power)/100000)
		}
	} else if string(charging) == "Not charging\n" {
		fmt.Println("Not Charging")
	}
}
