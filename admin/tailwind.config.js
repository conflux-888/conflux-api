/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        critical: "#FF0000",
        high: "#FF8C00",
        medium: "#E6C000",
        low: "#00FF00",
        ink: {
          DEFAULT: "#FFFFFF",
          muted: "rgba(255,255,255,0.6)",
          faint: "rgba(255,255,255,0.35)",
        },
        surface: {
          DEFAULT: "rgba(128,128,128,0.06)",
          raised: "rgba(128,128,128,0.1)",
          border: "rgba(255,255,255,0.08)",
        },
      },
      backgroundImage: {
        "app-gradient": "linear-gradient(180deg, #0D0D26 0%, #1A1A40 100%)",
      },
      borderRadius: {
        card: "14px",
        btn: "12px",
      },
      fontFamily: {
        sans: [
          "-apple-system",
          "BlinkMacSystemFont",
          "Inter",
          "Segoe UI",
          "Roboto",
          "sans-serif",
        ],
      },
    },
  },
  plugins: [],
};
