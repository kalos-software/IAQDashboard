import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	build: {
		// Configure build options
		minify: 'terser', // Use terser for better minification
		terserOptions: {
			compress: {
				drop_console: true, // Remove console.log statements
			}
		},
		rollupOptions: {
			output: {
				manualChunks: {
					// Split vendor code into separate chunks - only include @unovis/svelte
					vendor: ['@unovis/svelte']
				}
			}
		},
		chunkSizeWarningLimit: 1000 // Increase chunk size warning limit (in KB)
	}
});
