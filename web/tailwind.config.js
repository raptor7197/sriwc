/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        comic: {
          bg: '#090a12',
          panel: '#101320',
          panelAlt: '#161a2e',
          card: '#0f121e',
          border: '#252c48',
          ink: '#05060a',
          red: '#ff2453',
          redDark: '#cf153d',
          blue: '#00e5ff',
          blueDark: '#009bb3',
          yellow: '#ffe600',
          yellowDark: '#d4bf00',
          paper: '#f3f5fa',
          muted: '#808ba8',
        },
      },
      fontFamily: {
        comic: ['Bangers', 'Impact', 'Arial Black', 'sans-serif'],
        display: ['"Chakra Petch"', '"Space Grotesk"', 'sans-serif'],
        sans: ['"Space Grotesk"', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        'comic': '4px 4px 0px 0px #000000',
        'comic-sm': '2px 2px 0px 0px #000000',
        'comic-lg': '6px 6px 0px 0px #000000',
        'comic-red': '4px 4px 0px 0px #ff2453',
        'comic-blue': '4px 4px 0px 0px #00e5ff',
        'comic-yellow': '4px 4px 0px 0px #ffe600',
      },
    },
  },
  plugins: [],
};
