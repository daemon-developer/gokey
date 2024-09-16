package main

/* WORK IN PROGRESS CODE


import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Pastel colors inspired by Catppuccin
	green  = lipgloss.Color("#a6e3a1")
	red    = lipgloss.Color("#f38ba8")
	orange = lipgloss.Color("#fab387")
	purple = lipgloss.Color("#cba6f7")
	cyan   = lipgloss.Color("#89dceb")
	grey   = lipgloss.Color("#6c7086")

	bracketStyle = lipgloss.NewStyle().Foreground(grey)
	letterStyle  = lipgloss.NewStyle().Foreground(green)
	numberStyle  = lipgloss.NewStyle().Foreground(purple)
	otherStyle   = lipgloss.NewStyle().Foreground(cyan)
)

func (layout *Layout) String() string {
	return layout.stringInternal(false)
}

func (layout *Layout) StringWithCosts() string {
	return layout.stringInternal(true)
}

func (layout *Layout) stringInternal(costs bool) string {
	var sb strings.Builder

	// Write the layout name
	sb.WriteString(fmt.Sprintf("Layout: %s\n\n", layout.Name))

	// Determine the maximum number of rows
	maxRows := len(layout.Left.Rows)
	if len(layout.Right.Rows) > maxRows {
		maxRows = len(layout.Right.Rows)
	}

	// Generate the layout visualization
	for i := 0; i < maxRows; i++ {
		leftRow := ""
		rightRow := ""

		if i < len(layout.Left.Rows) {
			leftRow += visualizeRow(layout.Left.Rows[i], costs)
		}
		if i < len(layout.Right.Rows) {
			rightRow += visualizeRow(layout.Right.Rows[i], costs)
		}

		if !costs {
			sb.WriteString(fmt.Sprintf("%25s  |  %s\n", leftRow, rightRow))
		} else {
			sb.WriteString(fmt.Sprintf("%42s  |  %s\n", leftRow, rightRow))
		}
	}

	return sb.String()
}

func visualizeRow(row []KeyPhysicalInfo, costs bool) string {
	var keys []string

	if !costs {
		for _, keyInfo := range row {
			if keyInfo.key.UnshiftedRune != 0 {
				displayRune := RuneDisplayVersion(unicode.ToUpper(keyInfo.key.UnshiftedRune))
				char := string(displayRune)
				var styledChar string
				if unicode.IsLetter(displayRune) {
					styledChar = letterStyle.Render(char)
				} else if unicode.IsNumber(displayRune) {
					styledChar = numberStyle.Render(char)
				} else {
					styledChar = otherStyle.Render(char)
				}
				keys = append(keys, fmt.Sprintf("%s%s%s", bracketStyle.Render("["), styledChar, bracketStyle.Render("]")))
			} else {
				keys = append(keys, fmt.Sprintf("%s %s", bracketStyle.Render("["), bracketStyle.Render("]")))
			}
		}
	} else {
		for _, keyInfo := range row {
			cost := keyInfo.cost
			costColor := lipgloss.Color(getHeatMapColor(cost))
			costStyle := lipgloss.NewStyle().Foreground(costColor)
			keys = append(keys, fmt.Sprintf("%s%s%s", bracketStyle.Render("["), costStyle.Render(fmt.Sprintf("%1.2f", cost)), bracketStyle.Render("]")))
		}
	}

	return strings.Join(keys, " ")
}

func getHeatMapColor(cost float64) string {
	// Normalize cost to 0-1 range
	normalizedCost := cost / 10.0
	if normalizedCost > 1 {
		normalizedCost = 1
	}

	// Interpolate between green, orange, and red
	if normalizedCost < 0.5 {
		return lipgloss.Color(green).Blend(lipgloss.Color(orange), normalizedCost*2).Hex()
	} else {
		return lipgloss.Color(orange).Blend(lipgloss.Color(red), (normalizedCost-0.5)*2).Hex()
	}
}
*/
