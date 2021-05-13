package cli

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
)

const (
	TerminalLineTypeTask = iota + 1
)

type TerminalLine struct {
	Type        int
	Description string
	Completed   bool
	Start       time.Time
	Duration    time.Duration

	spinner *pterm.SpinnerPrinter
}

type Terminal struct {
	lines []*TerminalLine
}

func NewTerminal() *Terminal {
	return &Terminal{
		lines: make([]*TerminalLine, 0),
	}
}

func (t *Terminal) Error(err error) {
	pterm.Error.WithShowLineNumber(false).Println(err.Error())
}

func (t *Terminal) Section(title string) {
	pterm.DefaultSection.Println(title)
}

func (t *Terminal) Box(text string) {
	pterm.DefaultBox.Println(text)
}

func (t *Terminal) LineBreak() {
	pterm.DefaultBasicText.Println()
}

func (t *Terminal) InProgressTask(description string) {
	sp, err := pterm.DefaultSpinner.Start(description)
	if err != nil {
		panic(err)
	}

	t.lines = append(t.lines, &TerminalLine{
		Type:        TerminalLineTypeTask,
		Description: description,
		Completed:   false,
		Start:       time.Now(),

		spinner: sp,
	})
}

func (t *Terminal) TaskSucceeded() {
	t.taskCompleted(nil)
}

func (t *Terminal) TaskFailed(err error) {
	t.taskCompleted(err)
}

func (t *Terminal) taskCompleted(err error) {
	for i := len(t.lines) - 1; i >= 0; i-- {
		task := t.lines[i]

		if task.Type == TerminalLineTypeTask && !task.Completed {
			task.Completed = true
			task.Duration = time.Now().Sub(task.Start).Round(time.Millisecond)

			if err == nil {
				task.spinner.Success(fmt.Sprintf("%s (Took %s)", task.Description, task.Duration.String()))
				continue
			}

			task.spinner.FailPrinter = pterm.Error.WithShowLineNumber(false)
			task.spinner.Fail(err.Error())
		}
	}
}
