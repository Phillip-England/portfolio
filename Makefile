dev:
	bun run --hot index.ts

tw:
	tailwindcss -i './static/input.css' -o './static/output.css' --watch

bundle:
	bun build './src/client/three.ts' --outfile './static/three.js' --watch