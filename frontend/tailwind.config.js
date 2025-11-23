/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: "#0F172A",   // Slate 900
        secondary: "#1E293B", // Slate 800
        tertiary: "#334155",  // Slate 700
        accent: {
          gold: "#F59E0B",    // Amber 500
          blue: "#3B82F6",    // Blue 500
        },
        text: {
          primary: "#F8FAFC",   // Slate 50
          secondary: "#94A3B8", // Slate 400
        },
        status: {
          success: "#10B981", // Emerald 500
          error: "#EF4444",   // Red 500
        }
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
