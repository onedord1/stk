package config

import (
	"github.com/gdamore/tcell/v2"
)

// Theme defines colors for the UI
type Theme struct {
	Name       string
	Background tcell.Color
	Foreground tcell.Color
	Primary    tcell.Color
	Secondary  tcell.Color
	Success    tcell.Color
	Warning    tcell.Color
	Error      tcell.Color
	Muted      tcell.Color
	Border     tcell.Color
	Highlight  tcell.Color
}

// GetTheme returns a theme by name
func GetTheme(name string) *Theme {
	themes := map[string]*Theme{
		"catppuccin":  Catppuccin(),
		"dracula":     Dracula(),
		"nord":        Nord(),
		"gruvbox":     Gruvbox(),
		"solarized":   Solarized(),
		"tokyo-night": TokyoNight(),
		"monokai":     Monokai(),
		"one-dark":    OneDark(),
		"cyberpunk":   Cyberpunk(),
		"forest":      Forest(),
		"ocean":       Ocean(),
		"sunset":      Sunset(),
	}

	if theme, ok := themes[name]; ok {
		return theme
	}
	return Catppuccin()
}

// Catppuccin theme (Mocha variant)
func Catppuccin() *Theme {
	return &Theme{
		Name:       "catppuccin",
		Background: tcell.NewRGBColor(30, 30, 46),
		Foreground: tcell.NewRGBColor(205, 214, 244),
		Primary:    tcell.NewRGBColor(137, 180, 250),
		Secondary:  tcell.NewRGBColor(180, 190, 254),
		Success:    tcell.NewRGBColor(166, 227, 161),
		Warning:    tcell.NewRGBColor(249, 226, 175),
		Error:      tcell.NewRGBColor(243, 139, 168),
		Muted:      tcell.NewRGBColor(69, 71, 90),
		Border:     tcell.NewRGBColor(88, 91, 112),
		Highlight:  tcell.NewRGBColor(203, 166, 247),
	}
}

// Dracula theme
func Dracula() *Theme {
	return &Theme{
		Name:       "dracula",
		Background: tcell.NewRGBColor(40, 42, 54),
		Foreground: tcell.NewRGBColor(248, 248, 242),
		Primary:    tcell.NewRGBColor(189, 147, 249),
		Secondary:  tcell.NewRGBColor(139, 233, 253),
		Success:    tcell.NewRGBColor(80, 250, 123),
		Warning:    tcell.NewRGBColor(241, 250, 140),
		Error:      tcell.NewRGBColor(255, 85, 85),
		Muted:      tcell.NewRGBColor(68, 71, 90),
		Border:     tcell.NewRGBColor(98, 114, 164),
		Highlight:  tcell.NewRGBColor(255, 121, 198),
	}
}

// Nord theme
func Nord() *Theme {
	return &Theme{
		Name:       "nord",
		Background: tcell.NewRGBColor(46, 52, 64),
		Foreground: tcell.NewRGBColor(236, 239, 244),
		Primary:    tcell.NewRGBColor(136, 192, 208),
		Secondary:  tcell.NewRGBColor(129, 161, 193),
		Success:    tcell.NewRGBColor(163, 190, 140),
		Warning:    tcell.NewRGBColor(235, 203, 139),
		Error:      tcell.NewRGBColor(191, 97, 106),
		Muted:      tcell.NewRGBColor(59, 66, 82),
		Border:     tcell.NewRGBColor(76, 86, 106),
		Highlight:  tcell.NewRGBColor(180, 142, 173),
	}
}

// Gruvbox theme (dark variant)
func Gruvbox() *Theme {
	return &Theme{
		Name:       "gruvbox",
		Background: tcell.NewRGBColor(40, 40, 40),
		Foreground: tcell.NewRGBColor(235, 219, 178),
		Primary:    tcell.NewRGBColor(131, 165, 152),
		Secondary:  tcell.NewRGBColor(211, 134, 155),
		Success:    tcell.NewRGBColor(184, 187, 38),
		Warning:    tcell.NewRGBColor(250, 189, 47),
		Error:      tcell.NewRGBColor(251, 73, 52),
		Muted:      tcell.NewRGBColor(60, 56, 54),
		Border:     tcell.NewRGBColor(80, 73, 69),
		Highlight:  tcell.NewRGBColor(254, 128, 25),
	}
}

// Solarized theme (dark variant)
func Solarized() *Theme {
	return &Theme{
		Name:       "solarized",
		Background: tcell.NewRGBColor(0, 43, 54),
		Foreground: tcell.NewRGBColor(131, 148, 150),
		Primary:    tcell.NewRGBColor(38, 139, 210),
		Secondary:  tcell.NewRGBColor(108, 113, 196),
		Success:    tcell.NewRGBColor(133, 153, 0),
		Warning:    tcell.NewRGBColor(181, 137, 0),
		Error:      tcell.NewRGBColor(220, 50, 47),
		Muted:      tcell.NewRGBColor(7, 54, 66),
		Border:     tcell.NewRGBColor(88, 110, 117),
		Highlight:  tcell.NewRGBColor(211, 54, 130),
	}
}

// TokyoNight theme
func TokyoNight() *Theme {
	return &Theme{
		Name:       "tokyo-night",
		Background: tcell.NewRGBColor(26, 27, 38),
		Foreground: tcell.NewRGBColor(169, 177, 214),
		Primary:    tcell.NewRGBColor(122, 162, 247),
		Secondary:  tcell.NewRGBColor(187, 154, 247),
		Success:    tcell.NewRGBColor(158, 206, 106),
		Warning:    tcell.NewRGBColor(224, 175, 104),
		Error:      tcell.NewRGBColor(247, 118, 142),
		Muted:      tcell.NewRGBColor(41, 46, 66),
		Border:     tcell.NewRGBColor(61, 66, 91),
		Highlight:  tcell.NewRGBColor(125, 207, 255),
	}
}

// Monokai theme
func Monokai() *Theme {
	return &Theme{
		Name:       "monokai",
		Background: tcell.NewRGBColor(39, 40, 34),
		Foreground: tcell.NewRGBColor(248, 248, 242),
		Primary:    tcell.NewRGBColor(102, 217, 239),
		Secondary:  tcell.NewRGBColor(174, 129, 255),
		Success:    tcell.NewRGBColor(166, 226, 46),
		Warning:    tcell.NewRGBColor(253, 151, 31),
		Error:      tcell.NewRGBColor(249, 38, 114),
		Muted:      tcell.NewRGBColor(60, 61, 54),
		Border:     tcell.NewRGBColor(73, 72, 62),
		Highlight:  tcell.NewRGBColor(230, 219, 116),
	}
}

// OneDark theme
func OneDark() *Theme {
	return &Theme{
		Name:       "one-dark",
		Background: tcell.NewRGBColor(40, 44, 52),
		Foreground: tcell.NewRGBColor(171, 178, 191),
		Primary:    tcell.NewRGBColor(97, 175, 239),
		Secondary:  tcell.NewRGBColor(198, 120, 221),
		Success:    tcell.NewRGBColor(152, 195, 121),
		Warning:    tcell.NewRGBColor(229, 192, 123),
		Error:      tcell.NewRGBColor(224, 108, 117),
		Muted:      tcell.NewRGBColor(52, 58, 68),
		Border:     tcell.NewRGBColor(62, 68, 81),
		Highlight:  tcell.NewRGBColor(86, 182, 194),
	}
}

// Cyberpunk theme
func Cyberpunk() *Theme {
	return &Theme{
		Name:       "cyberpunk",
		Background: tcell.NewRGBColor(13, 2, 33),
		Foreground: tcell.NewRGBColor(255, 255, 255),
		Primary:    tcell.NewRGBColor(0, 255, 255),
		Secondary:  tcell.NewRGBColor(255, 0, 255),
		Success:    tcell.NewRGBColor(0, 255, 136),
		Warning:    tcell.NewRGBColor(255, 236, 39),
		Error:      tcell.NewRGBColor(255, 38, 116),
		Muted:      tcell.NewRGBColor(35, 25, 65),
		Border:     tcell.NewRGBColor(128, 0, 255),
		Highlight:  tcell.NewRGBColor(255, 113, 206),
	}
}

// Forest theme
func Forest() *Theme {
	return &Theme{
		Name:       "forest",
		Background: tcell.NewRGBColor(22, 33, 27),
		Foreground: tcell.NewRGBColor(200, 215, 200),
		Primary:    tcell.NewRGBColor(80, 200, 120),
		Secondary:  tcell.NewRGBColor(100, 180, 140),
		Success:    tcell.NewRGBColor(50, 220, 90),
		Warning:    tcell.NewRGBColor(220, 180, 50),
		Error:      tcell.NewRGBColor(220, 80, 80),
		Muted:      tcell.NewRGBColor(35, 50, 40),
		Border:     tcell.NewRGBColor(50, 80, 55),
		Highlight:  tcell.NewRGBColor(150, 255, 150),
	}
}

// Ocean theme
func Ocean() *Theme {
	return &Theme{
		Name:       "ocean",
		Background: tcell.NewRGBColor(15, 25, 45),
		Foreground: tcell.NewRGBColor(200, 220, 240),
		Primary:    tcell.NewRGBColor(100, 180, 255),
		Secondary:  tcell.NewRGBColor(120, 150, 200),
		Success:    tcell.NewRGBColor(60, 220, 180),
		Warning:    tcell.NewRGBColor(255, 200, 100),
		Error:      tcell.NewRGBColor(255, 100, 120),
		Muted:      tcell.NewRGBColor(25, 40, 65),
		Border:     tcell.NewRGBColor(40, 70, 110),
		Highlight:  tcell.NewRGBColor(80, 200, 255),
	}
}

// Sunset theme
func Sunset() *Theme {
	return &Theme{
		Name:       "sunset",
		Background: tcell.NewRGBColor(40, 25, 35),
		Foreground: tcell.NewRGBColor(255, 230, 220),
		Primary:    tcell.NewRGBColor(255, 140, 100),
		Secondary:  tcell.NewRGBColor(255, 100, 150),
		Success:    tcell.NewRGBColor(180, 220, 100),
		Warning:    tcell.NewRGBColor(255, 200, 80),
		Error:      tcell.NewRGBColor(255, 80, 100),
		Muted:      tcell.NewRGBColor(55, 40, 50),
		Border:     tcell.NewRGBColor(80, 55, 70),
		Highlight:  tcell.NewRGBColor(255, 180, 120),
	}
}

// AvailableThemes returns list of available theme names
func AvailableThemes() []string {
	return []string{
		"catppuccin",
		"dracula",
		"nord",
		"gruvbox",
		"solarized",
		"tokyo-night",
		"monokai",
		"one-dark",
		"cyberpunk",
		"forest",
		"ocean",
		"sunset",
	}
}
