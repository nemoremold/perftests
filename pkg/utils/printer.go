package utils

import (
	"strings"

	"k8s.io/klog/v2"
)

func PrintEmptyLine(level klog.Level) {
	klog.V(level).Info("\n")
}

// PrintDivider prints a divider
func PrintDivider(level klog.Level, width int) {
	klog.V(level).Infof("+%*v+", width, strings.Repeat("-", width))
}

// PrintDivider prints a string in the middle
func PrintCenter(level klog.Level, width int, str string) {
	blanks := width - len(str)
	left, right := blanks>>1, blanks>>1
	if blanks%2 != 0 {
		right++
	}
	klog.V(level).Infof("|%*v%v%*v|", left, strings.Repeat(" ", left), str, right, strings.Repeat(" ", right))
}
