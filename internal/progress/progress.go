package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Bar represents a progress bar
type Bar struct {
	total     int
	current   int
	width     int
	mu        sync.Mutex
	startTime time.Time
	prefix    string
}

// NewBar creates a new progress bar
func NewBar(total int, prefix string) *Bar {
	return &Bar{
		total:     total,
		current:   0,
		width:     40,
		startTime: time.Now(),
		prefix:    prefix,
	}
}

// Increment increases the progress by 1
func (b *Bar) Increment() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current++
	b.render()
}

// SetCurrent sets the current progress value
func (b *Bar) SetCurrent(current int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = current
	b.render()
}

// Finish marks the progress as complete
func (b *Bar) Finish() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = b.total
	b.render()
	fmt.Println() // Add newline after completion
}

// render displays the progress bar
func (b *Bar) render() {
	if b.total == 0 {
		return
	}

	percentage := float64(b.current) / float64(b.total) * 100
	filled := int(float64(b.width) * float64(b.current) / float64(b.total))

	// Build progress bar
	bar := strings.Repeat("█", filled) + strings.Repeat("░", b.width-filled)

	// Calculate elapsed and estimated time
	elapsed := time.Since(b.startTime)
	var eta string
	if b.current > 0 {
		avgTimePerItem := elapsed / time.Duration(b.current)
		remaining := time.Duration(b.total-b.current) * avgTimePerItem
		eta = formatDuration(remaining)
	} else {
		eta = "--:--"
	}

	// Format the complete line
	line := fmt.Sprintf("\r%s [%s] %d/%d (%.1f%%) ETA: %s",
		b.prefix,
		color.CyanString(bar),
		b.current,
		b.total,
		percentage,
		eta)

	fmt.Print(line)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

// MultiBar manages multiple progress operations
type MultiBar struct {
	operations map[string]*OperationStatus
	mu         sync.Mutex
	active     bool
}

// OperationStatus represents the status of an ongoing operation
type OperationStatus struct {
	Name      string
	Status    string // "pending", "running", "completed", "failed"
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

// NewMultiBar creates a new multi-operation progress tracker
func NewMultiBar() *MultiBar {
	return &MultiBar{
		operations: make(map[string]*OperationStatus),
		active:     true,
	}
}

// AddOperation adds a new operation to track
func (mb *MultiBar) AddOperation(name string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.operations[name] = &OperationStatus{
		Name:   name,
		Status: "pending",
	}
	mb.render()
}

// StartOperation marks an operation as started
func (mb *MultiBar) StartOperation(name string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if op, exists := mb.operations[name]; exists {
		op.Status = "running"
		op.StartTime = time.Now()
	}
	mb.render()
}

// CompleteOperation marks an operation as completed
func (mb *MultiBar) CompleteOperation(name string, err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if op, exists := mb.operations[name]; exists {
		if err != nil {
			op.Status = "failed"
			op.Error = err
		} else {
			op.Status = "completed"
		}
		op.EndTime = time.Now()
	}
	mb.render()
}

// Finish completes the multi-bar display
func (mb *MultiBar) Finish() {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.active = false
	mb.render()
	fmt.Println() // Add final newline
}

// render displays the current status of all operations
func (mb *MultiBar) render() {
	if !mb.active {
		return
	}

	// Clear previous lines
	fmt.Print("\r\033[K") // Clear current line

	completed := 0
	failed := 0
	running := 0

	// Count statuses
	for _, op := range mb.operations {
		switch op.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "running":
			running++
		}
	}

	total := len(mb.operations)

	// Display summary
	fmt.Printf("Progress: %d/%d complete", completed, total)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	if running > 0 {
		fmt.Printf(" (%d running)", running)
	}

	// Show currently running operations
	if running > 0 {
		var runningOps []string
		for _, op := range mb.operations {
			if op.Status == "running" {
				runningOps = append(runningOps, op.Name)
			}
		}
		if len(runningOps) <= 3 {
			fmt.Printf(" - %s", strings.Join(runningOps, ", "))
		} else {
			fmt.Printf(" - %s and %d more", strings.Join(runningOps[:2], ", "), len(runningOps)-2)
		}
	}
}
