package reporter

import (
	"fmt"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

type Reporter struct{}

func New() *Reporter {
	return &Reporter{}
}

func (r *Reporter) WorkflowStarted(name string) {
	fmt.Printf("\n%s🚀 Starting workflow: %s%s%s\n", ColorBold, ColorBlue, name, ColorReset)
	fmt.Println(stringsRepeat("-", 40))
}

func (r *Reporter) StepStarted(name string, stepType string) {
	fmt.Printf("%s• Running %s %s[%s]%s... ", ColorGray, name, ColorBold, stepType, ColorReset)
}

func (r *Reporter) StepPassed(duration time.Duration) {
	fmt.Printf("%sPASS%s %s(%v)%s\n", ColorGreen, ColorReset, ColorGray, duration.Truncate(time.Millisecond), ColorReset)
}

func (r *Reporter) StepFailed(err error, duration time.Duration, ignored bool) {
	if ignored {
		fmt.Printf("%sIGNORED%s %s(%v)%s\n", ColorYellow, ColorReset, ColorGray, duration.Truncate(time.Millisecond), ColorReset)
		fmt.Printf("  %s⚠ %v%s\n", ColorYellow, err, ColorReset)
	} else {
		fmt.Printf("%sFAIL%s %s(%v)%s\n", ColorRed, ColorReset, ColorGray, duration.Truncate(time.Millisecond), ColorReset)
		fmt.Printf("\n%s  Error Details:%s\n", ColorRed, ColorReset)
		fmt.Printf("  %v\n\n", err)
	}
}

func (r *Reporter) WorkflowFinished(name string, success bool, duration time.Duration) {
	fmt.Println(stringsRepeat("-", 40))
	status := fmt.Sprintf("%sSUCCESS%s", ColorGreen, ColorReset)
	if !success {
		status = fmt.Sprintf("%sFAILED%s", ColorRed, ColorReset)
	}
	fmt.Printf("%s✨ Workflow %s finished: %s %s(total %v)%s\n\n", ColorBold, name, status, ColorGray, duration.Truncate(time.Millisecond), ColorReset)
}

func stringsRepeat(s string, n int) string {
	res := ""
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}
