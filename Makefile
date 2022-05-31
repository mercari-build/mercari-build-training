.PHONY: build image
build image:
	docker build -t build2022/app:latest -f python/dockerfile .

.PHONY: run image
run image:
	docker run -dp 9000:9000 build2022/app:latest

.PHONY: run app
run app:
	uvicorn main:app --reload --port 9000

# .PHONY: add-test-item-with-image
# add-test-item-with-image:
# 	curl -X POST --url http://localhost:9000/items -F name=new-item -F category=book -F image=@python/images/book.jpg

