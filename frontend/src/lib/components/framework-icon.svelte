<script lang="ts">
    import type { Framework } from '$lib/state/projects.svelte';

    interface Props {
        framework: Framework | string;
        class?: string;
    }

    let { framework, class: className = 'size-4 shrink-0' }: Props = $props();
</script>

{#if framework === 'gin'}
    <img src="/gin.png" alt="Gin" class={className} />
{:else if framework === 'fiber'}
    <svg class={className} viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="1" y="3" width="10" height="2.5" rx="0.5" fill="#00ACD7"/>
        <rect x="1" y="6.75" width="14" height="2.5" rx="0.5" fill="#00ACD7"/>
        <rect x="1" y="10.5" width="6" height="2.5" rx="0.5" fill="#00ACD7"/>
    </svg>
{:else if framework === 'chi'}
    <img src="/chi.png" alt="Chi" class={className} />
{:else if framework === 'fasthttp'}
    <img src="/fasthttp.png" alt="FastHTTP" class={className} />
{:else if framework === 'stdlib'}
    <img src="/stdlib.png" alt="Standard Library" class={className} />
{:else if framework === 'react'}
    <svg class={className} viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="8" cy="8" r="1.5" fill="#61DAFB"/>
        <ellipse cx="8" cy="8" rx="6.5" ry="2.5" stroke="#61DAFB" stroke-width="0.8" fill="none"/>
        <ellipse cx="8" cy="8" rx="6.5" ry="2.5" stroke="#61DAFB" stroke-width="0.8" fill="none" transform="rotate(60 8 8)"/>
        <ellipse cx="8" cy="8" rx="6.5" ry="2.5" stroke="#61DAFB" stroke-width="0.8" fill="none" transform="rotate(120 8 8)"/>
    </svg>
{:else if framework === 'svelte'}
    <img src="/svelte.png" alt="Svelte" class={className} />
{:else if framework === 'vuejs'}
    <svg class={className} viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M12.8 1H10.4L8 5L5.6 1H3.2L8 9.5L12.8 1Z" fill="#41B883"/>
        <path d="M15 1H12.8L8 9.5L3.2 1H1L8 13L15 1Z" fill="#41B883" opacity="0.6"/>
        <path d="M12.8 1H10.4L8 5L5.6 1H3.2L8 9.5L12.8 1Z" fill="#35495E"/>
    </svg>
{:else if framework === 'nextjs'}
    <img src="/nextjs.png" alt="Next.js" class={className} />
{:else if framework === 'nestjs'}
    <img src="/nestjs.png" alt="NestJS" class={className} />
{:else if framework === 'express'}
    <img src="/express.png" alt="Express" class={className} />
{:else if framework === 'remix'}
    <img src="/remix.png" alt="Remix" class={className} />
{:else if framework === 'cloudflare'}
    <img src="/cloudflare.png" alt="Cloudflare" class={className} />
{:else if framework === 'opentelemetry'}
    <img src="/opentelemetry.png" alt="OpenTelemetry" class={className} />
{:else if framework === 'symfony'}
    <svg class={className} viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="8" cy="8" r="7.5" fill="currentColor" class="text-foreground"/>
        <path d="M11.5 4.5C11 4 10 3.8 9.2 4.3C8.3 4.9 7.8 5.8 7.3 6.5L7 6.4C6.2 6.1 5.2 6 4.5 6.8C3.9 7.5 4.1 8.4 4.6 8.9L5.3 8.3C5 8 4.9 7.6 5.2 7.3C5.6 6.8 6.4 7 7 7.2L6 9C5.6 9.6 5 10.5 4.3 10.6C3.9 10.7 3.5 10.5 3.3 10.2L2.5 10.6C2.9 11.3 3.6 11.5 4.2 11.4C5.1 11.2 5.8 10.3 6.3 9.5L7.1 8.2L7.8 8.6C8.5 9 8.6 9.7 8.1 10.3L8.8 10.8C9.6 9.8 9.4 8.7 8.5 8.1L7.9 7.7L8 7.5C8.5 6.6 9 5.7 9.6 5.2C10.1 4.8 10.7 4.9 11 5.2L11.5 4.5Z" fill="currentColor" class="text-background"/>
    </svg>
{:else}
    <img src="/custom.png" alt="Custom" class={className} />
{/if}
