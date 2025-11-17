dev:
	bun run --hot index.ts

tw:
	tailwindcss -i './static/input.css' -o './static/output.css' --watch

bundle:
	bun build './src/client/three.ts' --outfile './static/three.js' --watch

docker-build:
	sudo docker build --progress=plain -t myapp .

docker-clean:
	sudo docker rmi myapp:latest

docker-run:
	sudo docker run \
		--restart=unless-stopped \
		-p 8080:8080 \
		-v myapp_db:/app/data \
		--name myapp_container \
		-d \
		myapp

