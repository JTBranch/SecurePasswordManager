unitTest:
    go test ./... -short

e2eTest:
    go test ./ui/e2e ./tests/e2e