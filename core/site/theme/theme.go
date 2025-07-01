package theme

// Theme represents a complete design system configuration
type Root struct {
	Name      string
	Base      string // "light" or "dark"
	Tokens    ThemeTokens
	Variables map[string]string
}

type ThemeTokens struct {
	Colors     ColorTokens
	Spacing    SpacingTokens
	Typography TypographyTokens
	Borders    BorderTokens
	Shadows    ShadowTokens
	Animations AnimationTokens
}

// Detailed token structures
type ColorTokens struct {
	Primary          string `json:"primary"`
	PrimaryContent   string `json:"primaryContent"`
	Secondary        string `json:"secondary"`
	SecondaryContent string `json:"secondaryContent"`
	Accent           string `json:"accent"`
	AccentContent    string `json:"accentContent"`
	Neutral          string `json:"neutral"`
	NeutralContent   string `json:"neutralContent"`
	Base100          string `json:"base100"`
	Base200          string `json:"base200"`
	Base300          string `json:"base300"`
	BaseContent      string `json:"baseContent"`
	Info             string `json:"info"`
	InfoContent      string `json:"infoContent"`
	Success          string `json:"success"`
	SuccessContent   string `json:"successContent"`
	Warning          string `json:"warning"`
	WarningContent   string `json:"warningContent"`
	Error            string `json:"error"`
	ErrorContent     string `json:"errorContent"`
}

type SpacingTokens struct {
	Selector string `json:"selector"`
	Field    string `json:"field"`
	Base     string `json:"base"`
	Sm       string `json:"sm"`
	Md       string `json:"md"`
	Lg       string `json:"lg"`
	Xl       string `json:"xl"`
}

type TypographyTokens struct {
	FontSans       string `json:"fontSans"`
	FontMono       string `json:"fontMono"`
	FontSerif      string `json:"fontSerif"`
	FontSize       string `json:"fontSize"`
	FontSizeSm     string `json:"fontSizeSm"`
	FontSizeMd     string `json:"fontSizeMd"`
	FontSizeLg     string `json:"fontSizeLg"`
	FontSizeXl     string `json:"fontSizeXl"`
	LineHeight     string `json:"lineHeight"`
	LineHeightSm   string `json:"lineHeightSm"`
	LineHeightMd   string `json:"lineHeightMd"`
	FontWeight     string `json:"fontWeight"`
	FontWeightMd   string `json:"fontWeightMd"`
	FontWeightBold string `json:"fontWeightBold"`
}

type BorderTokens struct {
	Width          string `json:"width"`
	RadiusSelector string `json:"radiusSelector"`
	RadiusField    string `json:"radiusField"`
	RadiusBox      string `json:"radiusBox"`
	RadiusRound    string `json:"radiusRound"`
	Style          string `json:"style"`
	StyleDashed    string `json:"styleDashed"`
}

type ShadowTokens struct {
	Base  string `json:"base"`
	Sm    string `json:"sm"`
	Md    string `json:"md"`
	Lg    string `json:"lg"`
	Xl    string `json:"xl"`
	Inner string `json:"inner"`
	Focus string `json:"focus"`
	None  string `json:"none"`
}

type AnimationTokens struct {
	DurationFast   string `json:"durationFast"`
	DurationNormal string `json:"durationNormal"`
	DurationSlow   string `json:"durationSlow"`
	FunctionEase   string `json:"functionEase"`
	FunctionLinear string `json:"functionLinear"`
	FunctionBounce string `json:"functionBounce"`
}
