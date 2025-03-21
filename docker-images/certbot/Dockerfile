FROM ubuntu:20.04
RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -y certbot python3-certbot-dns-google curl jq
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl
COPY certbot.sh certbot.sh
RUN chmod +x certbot.sh
CMD ./certbot.sh
