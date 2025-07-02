package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
)