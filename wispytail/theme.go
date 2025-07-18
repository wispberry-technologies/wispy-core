package wispytail

const CssReset = `
/* Base styles / Reset */
@layer base {
  *, ::before, ::after, ::backdrop, ::file-selector-button {
    box-sizing: border-box;
    border-width: 0;
    border-style: solid;
    border-color: var(--color-gray-200);
    margin: 0;
    padding: 0;
  }

  :host, html {
    line-height: 1.5;
    font-family: var(--font-sans, ui-sans-serif, system-ui, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji");
    -webkit-tap-highlight-color: transparent;
    -webkit-text-size-adjust: 100%;
    -moz-tab-size: 4;
    tab-size: 4;
    font-feature-settings: normal;
    font-variation-settings: normal;
  }

  body {
    margin: 0;
    line-height: inherit;
  }

  h1, h2, h3, h4, h5, h6 {
    font-size: inherit;
    font-weight: inherit;
  }

  a {
    color: inherit;
    text-decoration: inherit;
    -webkit-text-decoration: inherit;
  }

  hr {
    height: 0;
    color: inherit;
    border-top-width: 1px;
  }

  table {
    text-indent: 0;
    border-color: inherit;
    border-collapse: collapse;
  }

  ol, ul, menu {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  fieldset {
    margin: 0;
    padding: 0;
  }

  legend {
    padding: 0;
  }

  button, input, optgroup, select, textarea {
    font: inherit;
    font-feature-settings: inherit;
    font-variation-settings: inherit;
    font-size: 100%;
    font-weight: inherit;
    line-height: inherit;
    letter-spacing: inherit;
    color: inherit;
    background: transparent;
    background-image: none;
    margin: 0;
    padding: 0;
  }

  input:where(:not([type='button'], [type='reset'], [type='submit'])), select, textarea {
    border-width: 1px;
  }

  button, [type='button'], [type='reset'], [type='submit'], ::file-selector-button {
    appearance: button;
    -webkit-appearance: button;
  }

  button, select {
    text-transform: none;
  }

  button:focus, input:focus, optgroup:focus, select:focus, textarea:focus {
    outline: 2px solid transparent;
    outline-offset: 2px;
  }

  input::placeholder, textarea::placeholder {
    opacity: 1;
    color: var(--color-gray-400, color-mix(in oklch, currentColor 50%, #0000));
  }

  :-moz-focusring {
    outline: auto;
  }

  :-moz-ui-invalid {
    box-shadow: none;
  }

  ::-webkit-search-decoration {
    -webkit-appearance: none;
  }

  textarea {
    resize: vertical;
  }

  summary {
    cursor: pointer;
  }

  abbr:where([title]) {
    text-decoration: underline dotted;
  }

  img, svg, video, canvas, audio, iframe, embed, object {
    display: block;
    vertical-align: middle;
  }

  img, video {
    max-width: 100%;
    height: auto;
  }

  code, kbd, samp, pre {
    font-family: var(--font-mono, monospace);
    font-size: 1em;
  }

  small {
    font-size: 80%;
  }

  sub, sup {
    font-size: 75%;
    line-height: 0;
    position: relative;
    vertical-align: baseline;
  }

  sub {
    bottom: -0.25em;
  }

  sup {
    top: -0.5em;
  }
}
`

var DefaultCssTheme = CssReset + `
/* Pale Wisp Theme - Light theme for Wispy Core */
:root {
  color-scheme: light;
  /* Mode - Light */
  --color-body: oklch(98% 0.001 106.423);
  --color-base-100: oklch(98% 0.001 106.423);
  --color-base-200: oklch(94% 0.004 286.32);
  --color-base-300: oklch(92% 0.004 286.32);
  --color-base-content: oklch(21.15% 0.012 254.09);
  /* ACCENTS */
  --color-primary: oklch(17.184% 0.02972 356.728);
  --color-primary-content: oklch(0.6727 0.2226 9.75);
  --color-secondary: oklch(0.8489 0.181 122.03);
  --color-secondary-content: oklch(16.779% 0.00761 164.031);
  --color-accent: oklch(98% 0.031 120.757);
  --color-accent-content: oklch(26% 0 0);
  --color-neutral: oklch(92% 0.004 286.32);
  --color-neutral-content: oklch(14% 0.005 285.823);
  /*  */
  --color-info: oklch(70% 0.165 254.624);
  --color-info-content: oklch(28% 0.091 267.935);
  --color-success: oklch(84% 0.238 128.85);
  --color-success-content: oklch(27% 0.072 132.109);
  --color-warning: oklch(75% 0.183 55.934);
  --color-warning-content: oklch(26% 0.079 36.259);
  --color-error: oklch(71% 0.194 13.428);
  --color-error-content: oklch(27% 0.105 12.094);
  --radius-selector: 0.5rem;
  --radius-field: 0.5rem;
  --radius-box: 0.25rem;
  --size-selector: 0.25rem;
  --size-field: 0.25rem;
  --border: 1px;
  --depth: 0;
  --noise: 0;
  --radius-selector: 0rem;
  --radius-field: 0.25rem;
  --radius-box: 0.25rem;
  --size-selector: 0.25rem;
  --size-field: 0.25rem;
  --border: 1px;
  --depth: 0;
  --noise: 0;
}
:root:has(body[color-mode=dark]),
[data-theme=dark] {
  color-scheme: "dark";
  /* Mode - Dark */
  --color-body: oklch(21% 0.006 285.885);
  --color-base-100: oklch(21% 0.006 285.885);
  --color-base-200: oklch(14% 0.005 285.823);
  --color-base-300: oklch(14% 0 0);
  --color-base-content: oklch(96% 0.001 286.375);
  /* ACCENTS */
  --color-primary: oklch(84% 0.238 128.85);
  --color-primary-content: oklch(26% 0.065 152.934);
  --color-secondary: oklch(39% 0.095 152.535);
  --color-secondary-content: oklch(94% 0.028 342.258);
  --color-accent: oklch(62% 0.214 259.815);
  --color-accent-content: oklch(97% 0.014 254.604);
  --color-neutral: oklch(27% 0.006 286.033);
  --color-neutral-content: oklch(98% 0.001 106.423);
}
`
