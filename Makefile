build:
	cd config-api && yarn && yarn build
	cd ./Finicky && xcodebuild clean build CODE_SIGN_IDENTITY="" CODE_SIGNING_REQUIRED=NO

run:
	bash ./scripts/run.sh
