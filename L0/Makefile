all: build

build:
	@docker compose -f ./srcs/config/docker-compose.yml --env-file srcs/config/.env up -d --build

down:
	@docker compose -f ./srcs/config/docker-compose.yml --env-file srcs/config/.env down

re: down clean build

make-topic:
	@docker exec -it kafka_con /opt/kafka/bin/kafka-topics.sh \
  --create \
  --topic orders \
  --bootstrap-server localhost:9092 \
  --partitions 1 \
  --replication-factor 1 \
  --if-not-exists

  docker exec -it kafka_con /opt/kafka/bin/kafka-topics.sh \
  --create \
  --topic __consumer_offsets \
  --bootstrap-server localhost:9092 \
  --partitions 50 \
  --replication-factor 1 \
  --config cleanup.policy=compact

docs:
	cd ./srcs/requirements/app && swag init --dir ./cmd,./internal/handler --parseDependency --parseInternal --output ./docs
clean: down
	@docker system prune -a --force

fclean:
	@docker stop $$(docker ps -qa)
	@docker system prune --all --force --volumes


.PHONY	: all build down re clean fclean

# con_to_db:
# 	psql -h localhost -U user_db -d orders_db

#docker exec -it db_con psql -U user_db -d orders_db
# docker exec -it kafka_con bash

#/opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092

#/opt/kafka/bin/kafka-topics.sh \
  --create \
  --topic orders \
  --bootstrap-server localhost:9092 \
  --partitions 1 \
  --replication-factor 1 \
  --if-not-exists