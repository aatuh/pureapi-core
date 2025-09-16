package logging

import (
	"fmt"
	"strings"
)

// Println formats and prints a debug colored message into stdout.
//
// Parameters:
//   - messages The messages to print.
func Println(messages ...any) {
	fmt.Println(defaultLogOpts.LogLevelOpts.Debug.Color, messages)
}

// PrintlnBoard formats and prints a debug colored message into stdout with
// visible signage.
//
// Parameters:
//   - messages The messages to print.
func PrintlnBoard(messages ...any) {
	for range 5 {
		fmt.Println(
			defaultLogOpts.LogLevelOpts.Debug.Color, strings.Repeat("=", 40),
		)
	}
	fmt.Println(
		defaultLogOpts.LogLevelOpts.Debug.Color, messages, ANSICodeReset,
	)
}

// PrintlnJSON formats and prints a debug colored JSON message into stdout.
//
// Parameters:
//   - messages The messages to print.
func PrintlnJSON(messages ...any) {
	fmt.Println(
		defaultLogOpts.LogLevelOpts.Debug.Color,
		AnyToJSONString(messages),
		ANSICodeReset,
	)
}
