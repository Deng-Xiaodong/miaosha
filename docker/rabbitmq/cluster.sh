# bin/bash

# 将主结点的erlang_cookie复制到宿主机 ./cookie
#echo "从主结点复制cookie"
#docker cp rabbitmq1:/var/lib/rabbitmq/.erlang.cookie ./cookie

# 将cookie复制到从结点/var/lib/rabbitmq
#echo "复制cookie到从结点"
#docker cp ./cookie rabbitmq2:/var/lib/rabbitmq/.erlang.cookie
#docker cp ./cookie rabbitmq3:/var/lib/rabbitmq/.erlang.cookie


# #cluster
echo "加入集群"
docker exec  rabbitmq2 /bin/bash -c 'rabbitmqctl stop_app'
docker exec  rabbitmq2 /bin/bash -c 'rabbitmqctl reset'
docker exec  rabbitmq2 /bin/bash -c 'rabbitmqctl  join_cluster --ram rabbit@rabbitmq1'
docker exec  rabbitmq2 /bin/bash -c 'rabbitmqctl start_app'

docker exec  rabbitmq3 /bin/bash -c 'rabbitmqctl stop_app'
docker exec  rabbitmq3 /bin/bash -c 'rabbitmqctl reset'
docker exec  rabbitmq3 /bin/bash -c 'rabbitmqctl  join_cluster --ram rabbit@rabbitmq1'
docker exec  rabbitmq3 /bin/bash -c 'rabbitmqctl start_app'

# 查看集群状态
echo "查看集群状态"
docker exec  rabbitmq1 /bin/bash -c 'rabbitmqctl cluster_status'
