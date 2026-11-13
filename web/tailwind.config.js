/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "#09090b",
        surface: {
          1: "#121215",
          2: "#18181b",
          3: "#27272a",
          border: "#27272a",
        },
        accent: {
          blue: "#3b82f6",
          emerald: "#10b981",
          rose: "#f43f5e",
          amber: "#f59e0b",
          purple: "#a855f7",
        },
      },
      fontFamily: {
        mono: ["JetBrains Mono", "Fira Code", "monospace"],
        sans: ["Inter", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "sans-serif"],
      },
    },
  },
  plugins: [],
};
