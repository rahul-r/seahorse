/** @type {import('tailwindcss').Config} */
module.exports = {
  content: {
    files: ["templates/*.html"],
  },
  theme: {
    extend: {
      colors: {
        current: "rgb(var(--color-content1) / <alpha-value>)",
        bkg: {
          1: "rgb(var(--color-bkg1) / <alpha-value>)",
          2: "rgb(var(--color-bkg2) / <alpha-value>)",
        },
        border: "rgb(var(--color-border) / <alpha-value>)",
        content: {
          1: "rgb(var(--color-content1) / <alpha-value>)",
          2: "rgb(var(--color-content2) / <alpha-value>)",
        },
      },
      animation: {
        "spin-slower": "spin 35s ease infinite",
        "spin-slow": "spin 25s ease-in-out infinite reverse",
      },
    },
  },
  plugins: [],
  darkMode: 'selector',

}
