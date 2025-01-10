package main

import "github.com/charmbracelet/lipgloss"

// Below styles define the visual appearance of various TUI elements.

var (
	tabStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69")).
			Padding(0, 1)

	activeTabStyle = tabStyle.Copy().
			Foreground(lipgloss.Color("205")).
			Underline(true)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)
)
