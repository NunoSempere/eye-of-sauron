# Others
deps:
	go mod tidy
	go mod download
	go mod vendor

env:
	find sources/* -type d -exec cp .env {} \;

restart-sources:
	sudo systemctl restart galerts
	sudo systemctl restart gdelt
	sudo systemctl restart globalbiodefense
	sudo systemctl restart gmw 
	sudo systemctl restart hn 
	sudo systemctl restart potpourri 
	sudo systemctl restart wikinews
	# sudo systemctl restart anthropic
	# sudo systemctl restart openai

