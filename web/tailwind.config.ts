import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: '#09090B',
        card: '#111113',
        border: 'rgba(255,255,255,0.06)',
        foreground: '#FAFAFA',
        muted: '#A1A1AA',
        accent: '#7C3AED',
        cyan: '#06B6D4',
      },
      boxShadow: {
        glow: '0 0 0 1px rgba(124,58,237,0.24), 0 20px 60px rgba(124,58,237,0.18)',
        soft: '0 24px 80px rgba(0,0,0,0.45)',
      },
      borderRadius: {
        '2xl': '1.25rem',
        '3xl': '1.75rem',
      },
      backgroundImage: {
        'dashboard-grid': 'radial-gradient(circle at top, rgba(124,58,237,0.2), transparent 30%), linear-gradient(to bottom, rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(to right, rgba(255,255,255,0.03) 1px, transparent 1px)',
      },
      animation: {
        shimmer: 'shimmer 2s linear infinite',
        pulseGlow: 'pulseGlow 2.5s ease-in-out infinite',
      },
      keyframes: {
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        pulseGlow: {
          '0%, 100%': { boxShadow: '0 0 0 0 rgba(124,58,237,0.18)' },
          '50%': { boxShadow: '0 0 0 10px rgba(124,58,237,0)' },
        },
      },
    },
  },
  plugins: [],
};

export default config;
