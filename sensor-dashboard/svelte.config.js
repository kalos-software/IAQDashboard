import adapterStatic from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Using the static adapter to generate a pure static site with no Node.js dependencies
		adapter: adapterStatic({
			// Path to where the static site should be built
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: true
		}),
		// Default to single-page app mode 
		prerender: { handleMissingId: 'ignore' }
	}
};

export default config;
