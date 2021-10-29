# assuming following commands have been executed
# minikube start --driver=hyperkit
# minikube addons enable ingress

# check env variables
if [[ ${RSS_GOOGLE_AUTH_CLIENT_ID} == "" ]]; then
  echo "Client ID for google auth is not specified (RSS_GOOGLE_AUTH_CLIENT_ID)"
  exit 1
fi

if [[ ${RSS_GOOGLE_AUTH_CLIENT_SECRET} == "" ]]; then
  echo "Secret for google auth is not specified (RSS_GOOGLE_AUTH_CLIENT_SECRET)"
  exit 1
fi

# create namespace
kubectl create ns rss

cd ./k8s

# init google auth secret
google_auth_client_id=$(printf "%s"  "${RSS_GOOGLE_AUTH_CLIENT_ID}" | base64)
google_auth_secret=$(printf "%s" "${RSS_GOOGLE_AUTH_CLIENT_SECRET}" | base64)

rm -f ./rss-secret-google-auth.yaml
cp ./rss-secret-google-auth-template.yaml ./rss-secret-google-auth.yaml

if [[ $OSTYPE == 'darwin'* ]]; then
  sed -i '' "s/client_id_template/$google_auth_client_id/" rss-secret-google-auth.yaml
  sed -i '' "s/secret_template/$google_auth_secret/" rss-secret-google-auth.yaml
else
  sed -i "s/client_id_template/$google_auth_client_id/" rss-secret-google-auth.yaml
  sed -i "s/secret_template/$google_auth_secret/" rss-secret-google-auth.yaml
fi

kubectl apply -f rss-secret-google-auth.yaml

# init db secret
rm -f ./rss-secret-db.yaml

db_username=$(printf "%s" "${POSTGRES_USER:-postgres}" | base64)
db_password=$(printf "%s" "${POSTGRES_PASSWORD:-postgres}" | base64)

rm -f ./rss-secret-db.yaml
cp ./rss-secret-db-template.yaml ./rss-secret-db.yaml

if [[ $OSTYPE == 'darwin'* ]]; then
  sed -i '' "s/username_template/$db_username/" rss-secret-db.yaml
  sed -i '' "s/password_template/$db_password/" rss-secret-db.yaml
else
  sed -i "s/username_template/$db_username/" rss-secret-db.yaml
  sed -i "s/password_template/$db_password/" rss-secret-db.yaml
fi

kubectl apply -f rss-secret-db.yaml

# apply configmap
kubectl apply -f rss-configmap.yaml

# set up database
kubectl create configmap init-script.sql --from-file=../db/sql/db.sql -n rss
kubectl apply -f postgres-pv.yaml
kubectl apply -f postgres-pvc.yaml
kubectl apply -f postgres-service.yaml
kubectl apply -f postgres-statefulset.yaml

# build service image
minikube image build --tag service-rss ..

# set up service
kubectl apply -f rss-deployment.yaml
kubectl apply -f rss-service.yaml

# set up ingress
kubectl apply -f rss-secret-tls.yaml
kubectl apply -f rss-ingress.yaml

# update /etc/hosts
echo
echo "sudo access required for modifying /etc/hosts"

# remove previous entries
if [[ $OSTYPE == 'darwin'* ]]; then
  sudo sed -i '' "/rss.aggregator.test.com/d" /etc/hosts
else
  sudo sed -i "/rss.aggregator.test.com/d" /etc/hosts
fi

sudo bash -c 'echo "$(minikube ip) rss.aggregator.test.com" >> /etc/hosts'