package service

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type StatsPoint struct {
	Timestamp int64   `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	MEM       uint64  `json:"mem"`
	NetUp     uint64  `json:"netUp"`
	NetDown   uint64  `json:"netDown"`
}

type StatsService struct {
	history    []StatsPoint
	maxPoints  int
	lock       sync.RWMutex
	prevNetIn  uint64
	prevNetOut uint64
}

var (
	instance *StatsService
	once     sync.Once
)

func GetStatsService() *StatsService {
	once.Do(func() {
		instance = &StatsService{
			history:   make([]StatsPoint, 0),
			maxPoints: 60, // Store last 60 points (1 minute if every 1s)
		}
		// Initialize network counters
		n, _ := net.IOCounters(false)
		if len(n) > 0 {
			instance.prevNetIn = n[0].BytesRecv
			instance.prevNetOut = n[0].BytesSent
		}
	})
	return instance
}

func (s *StatsService) Start() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			s.collect()
		}
	}()
}

func (s *StatsService) collect() {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	n, _ := net.IOCounters(false)

	var lastIn, lastOut uint64
	if len(n) > 0 {
		lastIn = n[0].BytesRecv - s.prevNetIn
		lastOut = n[0].BytesSent - s.prevNetOut
		s.prevNetIn = n[0].BytesRecv
		s.prevNetOut = n[0].BytesSent
	}

	var cpuPercent float64
	if len(c) > 0 {
		cpuPercent = c[0]
	}

	point := StatsPoint{
		Timestamp: time.Now().Unix(),
		CPU:       cpuPercent,
		MEM:       v.Used,
		NetUp:     lastOut,
		NetDown:   lastIn,
	}

	s.lock.Lock()
	s.history = append(s.history, point)
	if len(s.history) > s.maxPoints {
		s.history = s.history[1:]
	}
	s.lock.Unlock()
}

func (s *StatsService) GetHistory() []StatsPoint {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.history
}
