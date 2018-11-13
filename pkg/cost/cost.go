package cost

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"

	"managedkube.com/kube-cost-agent/pkg/node"
)

type PodCost struct {
	MinuteMemory float64
	HourMemory   float64
	DayMemory    float64
	MonthMemory  float64
	MinuteCpu    float64
	HourCpu      float64
	DayCpu       float64
	MonthCpu     float64
}

func Hello() {
	fmt.Println("Hello there")
}

func CalculatePodCost(node node.NodeInfo, podUsageMemory int64, podUsageCpu int64) PodCost {

	cost := PodCost{}

	computeCostPerHourMemory := node.ComputeCostPerHour * 0.5
	computeCostPerHourCpu := node.ComputeCostPerHour * 0.5

	percentUsedMemory := float64(podUsageMemory) / float64(node.MemoryCapacity)
	percentUsedCpu := float64(podUsageCpu) / float64(node.CpuCapacity)

	cost.HourMemory = computeCostPerHourMemory * float64(percentUsedMemory)
	cost.HourCpu = computeCostPerHourCpu * float64(percentUsedCpu)

	cost.MinuteMemory = cost.HourMemory / 60
	cost.MinuteCpu = cost.HourCpu / 60

	cost.DayMemory = cost.HourMemory * 24
	cost.DayCpu = cost.HourCpu * 24

	cost.MonthMemory = cost.DayMemory * 30
	cost.MonthCpu = cost.DayCpu * 30

	glog.V(3).Infof("Cost per minute memory: %s", strconv.FormatFloat(cost.MinuteMemory, 'f', 6, 64))
	glog.V(3).Infof("Cost per minute cpu: %s", strconv.FormatFloat(cost.MinuteCpu, 'f', 6, 64))

	glog.V(3).Infof("Cost per hour memory: %s", strconv.FormatFloat(cost.HourMemory, 'f', 6, 64))
	glog.V(3).Infof("Cost per hour cpu: %s", strconv.FormatFloat(cost.HourCpu, 'f', 6, 64))

	glog.V(3).Infof("Cost per day memory: %s", strconv.FormatFloat(cost.DayMemory, 'f', 6, 64))
	glog.V(3).Infof("Cost per day cpu: %s", strconv.FormatFloat(cost.DayCpu, 'f', 6, 64))

	glog.V(3).Infof("Cost per month memory: %s", strconv.FormatFloat(cost.MonthMemory, 'f', 6, 64))
	glog.V(3).Infof("Cost per month cpu: %s", strconv.FormatFloat(cost.MonthCpu, 'f', 6, 64))

	return cost
}
