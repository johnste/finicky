(
    cd packages/config-api
    npm ci
)

(
    cd packages/finicky-ui
    npm ci
)


(
    cd apps/finicky/src
    go mod tidy
)

