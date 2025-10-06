package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

var (
	TitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Padding(0).Margin(0)
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Padding(0).Margin(0)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0).Margin(0)
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Padding(0).Margin(0)
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Padding(0).Margin(0)

	infoColor    = color.New(color.FgBlue)
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed)
	titleColor   = color.New(color.FgMagenta, color.Bold)
	warningColor = color.New(color.FgYellow)
)

// Infof prints a formatted info message in blue
func Infof(format string, args ...interface{}) {
	infoColor.Printf(format+"\n", args...)
}

// Successf prints a formatted success message in green
func Successf(format string, args ...interface{}) {
	successColor.Printf(format+"\n", args...)
}

// Errorf prints a formatted error message in red
func Errorf(format string, args ...interface{}) {
	errorColor.Printf(format+"\n", args...)
}

// Info prints a message in blue
func Info(msg string) {
	infoColor.Println(msg)
}

// Success prints a message in green
func Success(msg string) {
	successColor.Println(msg)
}

// Error prints a message in red
func Error(msg string) {
	errorColor.Println(msg)
}

// Titlef prints a formatted title message in magenta bold
func Titlef(format string, args ...interface{}) {
	titleColor.Printf(format+"\n", args...)
}

// Title prints a title message in magenta bold
func Title(msg string) {
	titleColor.Println(msg)
}

// Warningf prints a formatted warning message in yellow
func Warningf(format string, args ...interface{}) {
	warningColor.Printf(format+"\n", args...)
}

// Warning prints a warning message in yellow
func Warning(msg string) {
	warningColor.Println(msg)
}
