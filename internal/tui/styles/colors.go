package styles

import "github.com/charmbracelet/lipgloss"

var (
	LinearPurple           = lipgloss.Color("#5E63D7")
	LinearLightPurple      = lipgloss.Color("#B8BAF9")
	LinearMediumPurple     = lipgloss.Color("#696EF0")
	LinearMutedPurple      = lipgloss.Color("#535470")
	LinearDarkMediumPurple = lipgloss.Color("#7075FF")

	Grey       = lipgloss.AdaptiveColor{Dark: "#868DB5", Light: "#2A2A38"}
	PurpleGrey = lipgloss.AdaptiveColor{Dark: "#606582", Light: "#797FA3"}
	White      = lipgloss.Color("#ffffff")

	Primary = lipgloss.AdaptiveColor{
		Dark:  string(LinearDarkMediumPurple),
		Light: string(LinearDarkMediumPurple),
	}

	Secondary = lipgloss.AdaptiveColor{
		Dark:  string(PurpleGrey.Dark),
		Light: string(PurpleGrey.Light),
	}

	Tertiary = lipgloss.AdaptiveColor{
		Dark:  string(Grey.Dark),
		Light: string(Grey.Light),
	}

	OverlayBG = lipgloss.AdaptiveColor{
		Dark:  string(LinearPurple),
		Light: string(LinearMutedPurple),
	}

	OverlayText = lipgloss.AdaptiveColor{
		Dark:  string(White),
		Light: string(LinearLightPurple),
	}

	OverlayTextHighlighted = lipgloss.AdaptiveColor{
		Dark:  string(Grey.Light),
		Light: string(White),
	}

	Green = lipgloss.AdaptiveColor{
		Dark:  "#77DD77",
		Light: "#56AE57",
	}

	Orange = lipgloss.AdaptiveColor{
		Dark:  "#FFB347",
		Light: "#EE7600",
	}
)
