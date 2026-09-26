.PHONY: setup doctor dev api worker check test build infra-up infra-down
setup:
	npm run setup
doctor:
	npm run doctor
dev:
	npm run dev
api:
	npm run api
worker:
	npm run worker
check:
	npm run check
test:
	npm test
build:
	npm run build
infra-up:
	npm run infra:up
infra-down:
	npm run infra:down
