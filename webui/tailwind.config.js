/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                border: "var(--color-border)",
                input: "var(--color-input)",
                ring: "var(--color-ring)",
                background: "var(--color-background)",
                foreground: "var(--color-foreground)",
                primary: {
                    DEFAULT: "var(--color-primary)",
                    foreground: "var(--color-primary-foreground)",
                },
                secondary: {
                    DEFAULT: "var(--color-secondary)",
                    foreground: "var(--color-secondary-foreground)",
                },
                destructive: {
                    DEFAULT: "var(--color-destructive)",
                    foreground: "var(--color-destructive-foreground)",
                },
                muted: {
                    DEFAULT: "var(--color-muted)",
                    foreground: "var(--color-muted-foreground)",
                },
                accent: {
                    DEFAULT: "var(--color-accent)",
                    foreground: "var(--color-accent-foreground)",
                },
                popover: {
                    DEFAULT: "var(--color-popover)",
                    foreground: "var(--color-popover-foreground)",
                },
                card: {
                    DEFAULT: "var(--color-card)",
                    foreground: "var(--color-card-foreground)",
                },
                // Semantic colors
                success: {
                    DEFAULT: "var(--ds-success)",
                    bg: "var(--ds-success-bg)",
                    border: "var(--ds-success-border)",
                },
                warning: {
                    DEFAULT: "var(--ds-warning)",
                    bg: "var(--ds-warning-bg)",
                    border: "var(--ds-warning-border)",
                },
                info: {
                    DEFAULT: "var(--ds-info)",
                    bg: "var(--ds-info-bg)",
                    border: "var(--ds-info-border)",
                },
                purple: {
                    DEFAULT: "var(--ds-purple)",
                    bg: "var(--ds-purple-bg)",
                    border: "var(--ds-purple-border)",
                },
                // Surface layers
                surface: {
                    DEFAULT: "var(--ds-surface)",
                    hover: "var(--ds-surface-hover)",
                },
            },
            borderRadius: {
                lg: "var(--radius)",
                md: "calc(var(--radius) - 2px)",
                sm: "calc(var(--radius) - 4px)",
                card: "var(--radius-card)",
                ctrl: "var(--radius-ctrl)",
                pill: "var(--radius-pill)",
            },
        },
    },
    plugins: [],
}