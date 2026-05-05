module.exports = {
    content: [
        "./app/**/*.{js,ts,jsx,tsx,mdx}",
        "./pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./components/**/*.{js,ts,jsx,tsx,mdx}",
        // Add this if you are using the src directory:
        "./src/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            fontFamily: {
                // This creates the 'font-sans' utility
                sans: ['var(--font-geist-sans)', 'ui-sans-serif', 'system-ui'],
                // This creates the 'font-mono' utility
                mono: ['var(--font-geist-mono)', 'ui-monospace', 'SFMono-Regular'],
            },
        },
    },
    // corePlugins: {
    //     preflight: false,
    // }
}