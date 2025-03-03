up:
	docker-compose up --build -d 

down:
	docker-compose down 

down-v:
	docker-compose down -v 


restart:
	docker-compose down && docker-compose up --build -d 

ps:
	docker ps 

logs:
	docker-compose logs -f
