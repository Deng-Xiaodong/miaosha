package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"time"
)

func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

func GetDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return diskInfo.UsedPercent
}

func main() {
	//fmt.Println(GetCpuPercent())
	//fmt.Println(GetMemPercent())
	//fmt.Println(GetDiskPercent())
	//info, _ := host.Info()
	//fmt.Println(info)
	//info, _ := cpu.Info() //总体信息
	//fmt.Println(info)
	//info1, _ := cpu.Times(false)
	//fmt.Println(info1)
	//info, _ := net.Connections("all") //可填入tcp、udp、tcp4、udp4等等
	//fmt.Println(info)
	info, _ := process.Pids() //获取当前所有进程的pid
	fmt.Println(info)

}
