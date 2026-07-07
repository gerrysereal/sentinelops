import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}', './lib/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        border: 'hsl(var(--border))',
        muted: 'hsl(var(--muted))',
        primary: 'hsl(var(--primary))',
        card: 'hsl(var(--card))'
      },
      boxShadow: {
        glow: '0 0 32px rgba(34, 211, 238, 0.14)'
      }
    }
  },
  plugins: [require('tailwindcss-animate')]
};

export default config;
