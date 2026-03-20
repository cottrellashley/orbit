import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Surfaces — 7-layer depth system
        bg: {
          DEFAULT: '#09090b',       // App shell (deepest)
          sidebar: '#0c0c0e',       // Sidebar panel
          content: '#0f0f12',       // Main content area (key layer)
          raised: '#141417',        // Cards, sections, elevated panels
          surface: '#19191d',       // Inputs, code blocks, inset areas
          hover: '#1e1e23',         // Hover state
          active: '#27272d',        // Active/pressed state
          overlay: '#0b0b0e',       // Modal/popover overlay tint
        },
        // Borders — translucent white for adaptability
        border: {
          DEFAULT: 'rgba(255, 255, 255, 0.08)',
          subtle: 'rgba(255, 255, 255, 0.05)',
          hover: 'rgba(255, 255, 255, 0.14)',
          active: 'rgba(255, 255, 255, 0.22)',
        },
        // Text — 4-tier luminance hierarchy
        txt: {
          DEFAULT: '#ececef',       // Primary — headings, important values
          secondary: '#a0a0ab',     // Secondary — body text, labels
          tertiary: '#63636e',      // Tertiary — timestamps, metadata
          quaternary: '#3f3f46',    // Quaternary — decorative, disabled
        },
        // Accent (Orange) — single brand color
        accent: {
          DEFAULT: '#F97316',
          soft: 'rgba(249, 115, 22, 0.10)',
          hover: '#FB923C',
          muted: 'rgba(249, 115, 22, 0.06)',
        },
        // Semantic — status colors only
        semantic: {
          green: '#34D399',
          'green-soft': 'rgba(52, 211, 153, 0.08)',
          yellow: '#FBBF24',
          'yellow-soft': 'rgba(251, 191, 36, 0.08)',
          red: '#F87171',
          'red-soft': 'rgba(248, 113, 113, 0.08)',
          orange: '#FB923C',
          'orange-soft': 'rgba(251, 146, 60, 0.08)',
          blue: '#60A5FA',
          'blue-soft': 'rgba(96, 165, 250, 0.08)',
        },
      },
      fontFamily: {
        sans: [
          'Inter',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'sans-serif',
        ],
        mono: [
          'SF Mono',
          'Cascadia Code',
          'Fira Code',
          'JetBrains Mono',
          'Consolas',
          'monospace',
        ],
      },
      fontSize: {
        // Type scale — strict sizes with line-height baked in
        'heading-xl': ['20px', { lineHeight: '28px', fontWeight: '600' }],
        'heading-lg': ['16px', { lineHeight: '24px', fontWeight: '600' }],
        'heading':    ['14px', { lineHeight: '20px', fontWeight: '600' }],
        'heading-sm': ['13px', { lineHeight: '18px', fontWeight: '600' }],
        'label-lg':   ['14px', { lineHeight: '20px', fontWeight: '500' }],
        'label':      ['13px', { lineHeight: '18px', fontWeight: '500' }],
        'label-sm':   ['12px', { lineHeight: '16px', fontWeight: '500' }],
        'body':       ['13px', { lineHeight: '20px', fontWeight: '400' }],
        'body-sm':    ['12px', { lineHeight: '18px', fontWeight: '400' }],
        'caption':    ['11px', { lineHeight: '16px', fontWeight: '500' }],
        'caption-xs': ['10px', { lineHeight: '14px', fontWeight: '500' }],
        'mono-label': ['12px', { lineHeight: '18px', fontWeight: '400' }],
        'mono-sm':    ['11px', { lineHeight: '16px', fontWeight: '400' }],
      },
      // Strict 4px grid spacing
      spacing: {
        '0.5': '2px',
        '1':   '4px',
        '1.5': '6px',
        '2':   '8px',
        '3':   '12px',
        '4':   '16px',
        '5':   '20px',
        '6':   '24px',
        '8':   '32px',
        '10':  '40px',
        '12':  '48px',
        '16':  '64px',
      },
      borderRadius: {
        sm: '4px',
        DEFAULT: '6px',
        lg: '8px',
        xl: '12px',
      },
      boxShadow: {
        sm:       '0 1px 2px rgba(0,0,0,0.3)',
        DEFAULT:  '0 1px 3px rgba(0,0,0,0.4), 0 0 0 1px rgba(255,255,255,0.03)',
        lg:       '0 4px 12px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.04)',
        xl:       '0 8px 24px rgba(0,0,0,0.6), 0 0 0 1px rgba(255,255,255,0.05)',
        'glow':   '0 0 0 1px rgba(249,115,22,0.15), 0 0 12px rgba(249,115,22,0.06)',
        'focus':  '0 0 0 2px rgba(249,115,22,0.20)',
      },
      transitionTimingFunction: {
        ease: 'cubic-bezier(0.4, 0, 0.2, 1)',
        'ease-out': 'cubic-bezier(0, 0, 0.2, 1)',
      },
      transitionDuration: {
        fast: '100ms',
        DEFAULT: '150ms',
        slow: '250ms',
      },
      animation: {
        'loading-slide': 'loading-slide 1.2s ease-in-out infinite',
        'pulse-dot': 'pulse-dot 2s ease-in-out infinite',
        spin: 'spin 0.7s linear infinite',
        'fade-in': 'fade-in 0.15s cubic-bezier(0, 0, 0.2, 1)',
        'modal-in': 'modal-in 0.2s cubic-bezier(0, 0, 0.2, 1)',
        'skeleton': 'skeleton 1.8s ease-in-out infinite',
        'page-in': 'page-in 0.12s cubic-bezier(0, 0, 0.2, 1)',
      },
      keyframes: {
        'loading-slide': {
          '0%':   { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(100%)' },
        },
        'pulse-dot': {
          '0%, 100%': { opacity: '1' },
          '50%':      { opacity: '0.3' },
        },
        'fade-in': {
          from: { opacity: '0' },
          to:   { opacity: '1' },
        },
        'modal-in': {
          from: { opacity: '0', transform: 'translateY(-6px) scale(0.98)' },
          to:   { opacity: '1', transform: 'translateY(0) scale(1)' },
        },
        'skeleton': {
          '0%':   { opacity: '0.04' },
          '50%':  { opacity: '0.08' },
          '100%': { opacity: '0.04' },
        },
        'page-in': {
          from: { opacity: '0', transform: 'translateY(2px)' },
          to:   { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
}

export default config
