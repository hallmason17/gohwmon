package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	logDir := filepath.Join(os.Getenv("HOME"), ".local", "state", "GoHWMon.log")
	logFile, err := os.OpenFile(logDir, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	defer logFile.Close()
	log.Println("Starting the application.")
	if err := ui.Init(); err != nil {
		log.Println("Error while initializing the UI.", err)
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
			log.Println("Closing via SIGTERM.")
			return
		case <-updateInt:
			go update(interval)
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				log.Println("Closing the application.")
				return
			}
		}
	}
}

func update(updateInterval time.Duration) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Println(err)
	}
	cpuPcnt, err := psCpu.Percent(updateInterval, true)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}

	fmt.Printf("Battery Cycle Count: %s", string(cycle_count))

	charging, err := os.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		log.Println(err)
	}
	energy_now, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_now")
	if err != nil {
		log.Println(err)
	}
	energy_now1 := string(energy_now)
	energy_now1 = energy_now1[:len(energy_now1)-2]
	energy, err := strconv.Atoi(string(energy_now1))
	if err != nil {
		log.Println(err)
	}

	capacity, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_full")
	capacity1 := string(capacity)
	capacity1 = capacity1[:len(capacity1)-2]
	capacity2, err := strconv.Atoi(string(capacity1))
	if err != nil {
		log.Println(err)
	}
	energy_full, err := os.ReadFile("/sys/class/power_supply/BAT0/energy_full_design")
	if err != nil {
		log.Println(err)
	}
	energy_full1 := string(energy_full)
	energy_full1 = energy_full1[:len(energy_full1)-2]
	max_energy, err := strconv.Atoi(energy_full1)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Battery Level: %.2f%%\n", float64(energy)/float64(capacity2)*100)

	if string(charging) != "Not charging\n" {
		power_now, err := os.ReadFile("/sys/class/power_supply/BAT0/power_now")
		if err != nil {
			log.Println(err)
		}
		power_now1 := string(power_now)
		power_now1 = power_now1[:len(power_now1)-2]
		power, err := strconv.Atoi(string(power_now1))
		if err != nil {
			log.Println(err)
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
