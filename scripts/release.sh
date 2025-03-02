mkdir -p dist
export $(cat .env | xargs) && gon scripts/gon-config.json

