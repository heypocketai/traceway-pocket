export const themeState = $state({
    isDark: true
});

export function initTheme() {
    if (typeof document !== 'undefined') {
        const stored = localStorage.getItem('theme');

        if (stored === 'dark' || stored === 'light') {
            themeState.isDark = stored === 'dark';
        } else {
            themeState.isDark = true;
        }

        document.documentElement.classList.toggle('dark', themeState.isDark);
        document.documentElement.style.colorScheme = themeState.isDark ? 'dark' : 'light';

        const observer = new MutationObserver(() => {
            themeState.isDark = document.documentElement.classList.contains('dark');
        });
        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class']
        });

        return () => {
            observer.disconnect();
        };
    }
}

export function toggleTheme() {
    themeState.isDark = !themeState.isDark;
    document.documentElement.classList.toggle('dark', themeState.isDark);
    document.documentElement.style.colorScheme = themeState.isDark ? 'dark' : 'light';
    localStorage.setItem('theme', themeState.isDark ? 'dark' : 'light');
}
