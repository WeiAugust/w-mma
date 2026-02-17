.PHONY: test test-e2e

test:
	cd backend && go test ./...

test-e2e:
	cd backend && GOPROXY=https://goproxy.cn,direct GOSUMDB=off go test ./tests/e2e -v
	cd admin && pnpm vitest run e2e/review_publish.spec.ts
	cd miniapp && npm test -- navigation.spec.js
