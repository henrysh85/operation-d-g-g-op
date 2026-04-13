/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx,js,jsx}'],
  theme: {
    extend: {
      colors: {
        // DCGG palette extracted from prototype v7
        brand: {
          50:  '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          500: '#2563eb',
          600: '#1d4ed8',
          700: '#1e40af',
        },
        ink: {
          900: '#111827',
          700: '#374151',
          500: '#6b7280',
          400: '#9ca3af',
          300: '#d1d5db',
          200: '#e5e7eb',
          100: '#f3f4f6',
          50:  '#f9fafb',
        },
        ok:    '#059669',
        warn:  '#d97706',
        err:   '#dc2626',
        info:  '#0891b2',
        plum:  '#7c3aed',
      },
      fontFamily: {
        sans: [
          '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'system-ui',
          'Helvetica', 'Arial', 'sans-serif',
        ],
      },
      fontSize: {
        xxs: ['11px', '14px'],
      },
      boxShadow: {
        card: '0 1px 2px rgba(17,24,39,0.04)',
      },
    },
  },
  plugins: [],
};
