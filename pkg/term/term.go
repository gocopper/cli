package term

import (
	"fmt"
	"os"
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

func (t *Terminal) Error(msg string, err error) {
	var printer = pterm.Error.WithShowLineNumber(false)

	if err == nil {
		printer.Println(msg)
	} else {
		printer.Println(msg + " because \n" + err.Error())
	}
}

func (t *Terminal) Fatal(msg string, err error) {
	t.Error(msg, err)
	os.Exit(1)
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

func (t *Terminal) Text(txt string) {
	pterm.DefaultBasicText.Println(txt)
}

func (t *Terminal) Success(txt string) {
	pterm.Success.Println(txt)
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
	for i := len(t.lines) - 1; i >= 0; i-- {
		task := t.lines[i]

		if task.Type == TerminalLineTypeTask && !task.Completed {
			task.Completed = true
			task.Duration = time.Now().Sub(task.Start).Round(time.Millisecond)

			task.spinner.Success(fmt.Sprintf("%s (Took %s)", task.Description, task.Duration.String()))
		}
	}
}

func (t *Terminal) TaskFailed(err error) {
	for i := len(t.lines) - 1; i >= 0; i-- {
		task := t.lines[i]

		if task.Type == TerminalLineTypeTask && !task.Completed {
			task.Completed = true
			task.Duration = time.Now().Sub(task.Start).Round(time.Millisecond)

			task.spinner.FailPrinter = pterm.Error.WithShowLineNumber(false)
			task.spinner.Fail(fmt.Sprintf("%s (Took %s)", task.Description, task.Duration.String()))

			pterm.Error.
				WithPrefix(pterm.Prefix{
					Text:  "       ",
					Style: &pterm.ThemeDefault.ErrorPrefixStyle,
				}).
				WithShowLineNumber(false).
				Println(err.Error())
		}
	}
}
