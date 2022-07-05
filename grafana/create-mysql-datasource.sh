#!/bin/bash

SHELL_PATH=$(cd "$(dirname "$0")";pwd)

GRAFANA_WORKSPACE_FILE="grafana-workspace-info.json"
GRAFANA_DATASOURCE_INPUT_FILE="datasource-info.json"
GRAFANA_DATASOURCE_OUTPUT_FILE="mysql-datasource-info.json"

# Generate datasource file.
DS_NAME="$(jq -r .context.stackName ${SHELL_PATH}/../cdk.json)"
DS_HOST="$(jq -r .${DS_NAME}.host ${SHELL_PATH}/../cdk.out/${GRAFANA_DATASOURCE_INPUT_FILE})"
DS_DATABASE="$(jq -r .${DS_NAME}.database ${SHELL_PATH}/../cdk.out/${GRAFANA_DATASOURCE_INPUT_FILE})"
DS_USER="$(jq -r .${DS_NAME}.user ${SHELL_PATH}/../cdk.out/${GRAFANA_DATASOURCE_INPUT_FILE})"
DS_PASSWORD="$(jq -r .${DS_NAME}.password ${SHELL_PATH}/../cdk.out/${GRAFANA_DATASOURCE_INPUT_FILE})"

echo "DS_NAME: ${DS_NAME}"
echo "DS_HOST: ${DS_HOST}"
echo "DS_DATABASE: ${DS_DATABASE}"
echo "DS_USER: ${DS_USER}"
echo "DS_PASSWORD: ${DS_PASSWORD}"

jq -n --arg dsName ${DS_NAME} \
    --arg dsHost ${DS_HOST} \
    --arg dsDb ${DS_DATABASE} \
    --arg dsUsr ${DS_USER} \
    --arg dsPwd ${DS_PASSWORD} \
    '{
        "name": $dsName,
        "type": "mysql",
        "url": $dsHost,
        "database": $dsDb,
        "user": $dsUsr,
        "password": $dsPwd,
        "access": "direct",
        "isDefault": true
    }' > ${SHELL_PATH}/${GRAFANA_DATASOURCE_OUTPUT_FILE}

WORKSPACE_ID="$(jq -r .workspaceId ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE})"
DEPLOYMENT_REGION="$(jq -r .deploymentRegion ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE})"

KEY_NAME="grafana-api-key"
KEY_ROLE="ADMIN"
KEY_DURATION=3600

API_ENDPOINT="$(jq -r .workspaceEndpoint ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE})"
aws grafana delete-workspace-api-key --key-name ${KEY_NAME} --workspace-id ${WORKSPACE_ID} --region ${DEPLOYMENT_REGION} 2>/dev/null
API_KEY="$(aws grafana create-workspace-api-key \
            --key-name ${KEY_NAME} \
            --key-role ${KEY_ROLE} \
            --seconds-to-live ${KEY_DURATION} \
            --workspace-id ${WORKSPACE_ID} \
            --region ${DEPLOYMENT_REGION} \
            --output text \
            --query key)"

echo "API_KEY: ${API_KEY}"

# Construct grafana api request
REST_API_URL_CREATE_DS="https://${API_ENDPOINT}/api/datasources"

echo "REST_API_URL_CREATE_DS: ${REST_API_URL_CREATE_DS}"

curl -X POST -H "Content-Type: application/json" \
             -H "Accept: application/json" \
             -H "Authorization: Bearer ${API_KEY}" \
             ${REST_API_URL_CREATE_DS} \
             -d @${SHELL_PATH}/${GRAFANA_DATASOURCE_OUTPUT_FILE}
