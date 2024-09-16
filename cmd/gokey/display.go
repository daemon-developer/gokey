package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var (
	filledStyle lipgloss.Style
	emptyStyle  lipgloss.Style

	// Catppuccin Mocha colors
	green    = "#a6e3a1"
	red      = "#f38ba8"
	blue     = "#89b4fa"
	lavender = "#b4befe"
	peach    = "#fab387"
	yellow   = "#f9e2af"
	text     = "#cdd6f4"
	surface2 = "#585b70" // For brackets

	bracketStyle lipgloss.Style
	letterStyle  lipgloss.Style
	numberStyle  lipgloss.Style
	symbolStyle  lipgloss.Style
	otherStyle   lipgloss.Style
)

func init() {
	surface0 := "#313244" // Surface0 (darker background)

	filledStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(blue)).
		Background(lipgloss.Color(surface0))

	emptyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(surface0)).
		Background(lipgloss.Color(surface0))

	bracketStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(surface2))
	letterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(lavender))
	numberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(peach))
	symbolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(yellow))
	otherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(text))
}

func generateProgressBar(percent float64, length int) string {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	filled := int(float64(length) * percent / 100)
	partial := int((float64(length)*percent/100)*8) % 8

	var bar strings.Builder

	// Full blocks
	bar.WriteString(filledStyle.Render(strings.Repeat("█", filled)))

	// Partial block
	if filled < length {
		partialChars := []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}
		if partial > 0 {
			bar.WriteString(filledStyle.Render(string(partialChars[partial])))
			filled++
		}

		// Empty space
		bar.WriteString(emptyStyle.Render(strings.Repeat(" ", length-filled)))
	}

	return bar.String()
}

func (layout *Layout) stringInternal(costs bool) string {
	var sb strings.Builder

	// Write the layout name
	// Write the layout name separately
	layoutNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(blue))
	layoutName := layoutNameStyle.Render(fmt.Sprintf("Layout: %s", layout.Name))
	sb.WriteString(layoutName + "\n\n")

	// Determine the maximum number of rows
	maxRows := len(layout.Left.Rows)
	if len(layout.Right.Rows) > maxRows {
		maxRows = len(layout.Right.Rows)
	}

	// Determine the width of the left side
	leftWidth := 30
	if costs {
		leftWidth = 42
	}

	// Generate the layout visualization
	for i := 0; i < maxRows; i++ {
		leftRow := ""
		rightRow := ""

		if i < len(layout.Left.Rows) {
			leftRow = visualizeRow(layout.Left.Rows[i], costs)
		}
		if i < len(layout.Right.Rows) {
			rightRow = visualizeRow(layout.Right.Rows[i], costs)
		}

		// Right-align the left row and left-align the right row
		leftAligned := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Right).Render(leftRow)
		rightAligned := lipgloss.NewStyle().Render(rightRow)

		sb.WriteString(fmt.Sprintf("%s  |  %s\n", leftAligned, rightAligned))
	}

	return sb.String()
}

func visualizeRow(row []KeyPhysicalInfo, costs bool) string {
	var keys []string

	if !costs {
		for _, keyInfo := range row {
			if keyInfo.key.UnshiftedRune != 0 {
				displayRune := RuneDisplayVersion(unicode.ToUpper(keyInfo.key.UnshiftedRune))
				keys = append(keys, formatKey(displayRune))
			} else {
				keys = append(keys, formatKey(' '))
			}
		}
	} else {
		for _, keyInfo := range row {
			keys = append(keys, formatCost(keyInfo.cost))
		}
	}

	return strings.Join(keys, " ")
}

func formatKey(r rune) string {
	var style lipgloss.Style
	switch {
	case unicode.IsLetter(r):
		style = letterStyle
	case unicode.IsNumber(r):
		style = numberStyle
	case unicode.IsPunct(r):
		style = symbolStyle
	default:
		style = otherStyle
	}
	return bracketStyle.Render("[") + style.Render(string(r)) + bracketStyle.Render("]")
}

func formatCost(cost float64) string {
	// Create a heat map color based on the cost (0-10 scale)
	heatColor := lipgloss.Color(blendColors(green, red, cost/10))
	costStyle := lipgloss.NewStyle().Foreground(heatColor)
	return bracketStyle.Render("[") + costStyle.Render(fmt.Sprintf("%1.2f", cost)) + bracketStyle.Render("]")
}

func blendColors(startColor, endColor string, ratio float64) string {
	// Simple linear interpolation between two colors
	start := parseHexColor(startColor)
	end := parseHexColor(endColor)

	blended := [3]int{
		int(float64(start[0])*(1-ratio) + float64(end[0])*ratio),
		int(float64(start[1])*(1-ratio) + float64(end[1])*ratio),
		int(float64(start[2])*(1-ratio) + float64(end[2])*ratio),
	}

	return fmt.Sprintf("#%02x%02x%02x", blended[0], blended[1], blended[2])
}

func parseHexColor(hex string) [3]int {
	hex = strings.TrimPrefix(hex, "#")
	var rgb [3]int
	fmt.Sscanf(hex, "%02x%02x%02x", &rgb[0], &rgb[1], &rgb[2])
	return rgb
}
