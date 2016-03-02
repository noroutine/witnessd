if [ ! -d ${HOME}/google-cloud-sdk ]; then
    curl https://sdk.cloud.google.com | bash
fi
gcloud config set project dominion-p2p
gcloud auth activate-service-account --key-file gcp-credentials.json

gcloud -q preview app deploy app.yaml --promote