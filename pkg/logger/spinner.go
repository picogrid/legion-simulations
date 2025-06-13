package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Spinner represents an animated spinner for long-running operations
type Spinner struct {
	mu       sync.Mutex
	active   bool
	message  string
	frames   []string
	interval time.Duration
	stopChan chan struct{}
}

// Default spinner frames
var (
	SpinnerDots   = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine   = []string{"-", "\\", "|", "/"}
	SpinnerCircle = []string{"◐", "◓", "◑", "◒"}
	SpinnerSquare = []string{"◰", "◳", "◲", "◱"}
	SpinnerArrow  = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	SpinnerBounce = []string{"⠁", "⠂", "⠄", "⠂"}
)

// NewSpinner creates a new spinner with the default frames
func NewSpinner(message string) *Spinner {
	return NewSpinnerWithFrames(message, SpinnerDots)
}

// NewSpinnerWithFrames creates a new spinner with custom frames
func NewSpinnerWithFrames(message string, frames []string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   frames,
		interval: 100 * time.Millisecond,
		stopChan: make(chan struct{}),
	}
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.stopChan:
				// Clear the line
				fmt.Printf("\r%s\r", strings.Repeat(" ", len(s.message)+10))
				return
			default:
				frame := s.frames[i%len(s.frames)]
				if l, ok := defaultLogger.(*logger); ok && !l.noColor {
					fmt.Printf("\r%s%s%s %s", colorCyan, frame, colorReset, s.message)
				} else {
					fmt.Printf("\r%s %s", frame, s.message)
				}
				i++
				time.Sleep(s.interval)
			}
		}
	}()
}

// Stop stops the spinner and optionally displays a final message
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	close(s.stopChan)
	time.Sleep(50 * time.Millisecond) // Give time for the goroutine to clean up
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	Success(message)
}

// Error stops the spinner and shows an error message
func (s *Spinner) Error(message string) {
	s.Stop()
	Error(message)
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// WithSpinner runs a function with a spinner
func WithSpinner(message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	err := fn()

	if err != nil {
		spinner.Error(fmt.Sprintf("%s failed: %v", message, err))
	} else {
		spinner.Success(fmt.Sprintf("%s completed", message))
	}

	return err
}

// ProgressBar represents a simple progress bar
type ProgressBar struct {
	total   int
	current int
	width   int
	message string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		total:   total,
		current: 0,
		width:   40,
		message: message,
	}
}

// Update updates the progress bar
func (p *ProgressBar) Update(current int) {
	p.current = current
	p.draw()
}

// Increment increments the progress bar by 1
func (p *ProgressBar) Increment() {
	p.current++
	p.draw()
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	p.current = p.total
	p.draw()
	fmt.Println()
}

func (p *ProgressBar) draw() {
	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	if l, ok := defaultLogger.(*logger); ok && !l.noColor {
		fmt.Printf("\r%s: %s%s%s %3.0f%%",
			p.message,
			colorGreen, bar, colorReset,
			percent*100)
	} else {
		fmt.Printf("\r%s: [%s] %3.0f%%",
			p.message,
			bar,
			percent*100)
	}
}
