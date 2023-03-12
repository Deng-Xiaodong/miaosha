docker-compose down
cd src/miaosha
git pull
cd ../../
docker start miaosha-complier
docker-compose up redis rabbitmq1 rabbitmq2 rabbitmq3 -d
sleep 10
./cluster.sh
docker-compose up publish1 publish2 publish3 consumer -d
sleep 10
docker-compose up nginx -d