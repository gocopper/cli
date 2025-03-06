package notifier

import "os/exec"

import "time"

const NotificationGroup = "copper"

type NotifyParams struct {
	Message string

	Title       string
	RemoveAfter time.Duration
}

func Notify(p NotifyParams) {
	if p.Message == "" {
		return
	}

	cmdArgs := []string{
		"-group", NotificationGroup,
		"-message", p.Message,
	}

	if p.Title != "" {
		cmdArgs = append(cmdArgs, "-title", p.Title)
	}
	
	cmd := exec.Command("terminal-notifier", cmdArgs...)
	defer func() { _ = cmd.Wait() }()

	_ = cmd.Start()

	if p.RemoveAfter > 0 {
		go func() {
			time.Sleep(p.RemoveAfter)
			removeCmd := exec.Command("terminal-notifier", "-remove", NotificationGroup)
			defer func() { _ = removeCmd.Wait() }()
			_ = removeCmd.Start()
		}()
	}
}
