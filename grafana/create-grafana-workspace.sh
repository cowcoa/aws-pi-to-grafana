#!/bin/bash

arg_count=$#
script_name=$(basename $0)
script_action=create

if test $arg_count -eq 1; then
  if [[ $1 =~ ^(create|delete)$ ]]; then
    script_action=$1
  else
    echo "Script Action must be create or delete"
    echo "Usage: $script_name [create|delete]"
    exit -1
  fi
else
  echo "Usage: $script_name [create|delete]"
  echo ""
  echo "Examples:"
  echo "$script_name create"
  echo ""
  exit 0
fi

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)
GRAFANA_WORKSPACE_FILE="grafana-workspace-info.json"
GRAFANA_DATASOURCE_FILE="mysql-datasource-info.json"

createWorkspace()
{
    local WORKSPACE_NAME="$(jq -r .context.stackName ${SHELL_PATH}/../cdk.json)"
    local ROLE_NAME="${WORKSPACE_NAME}-Role"
    local DEPLOYMENT_REGION="$(jq -r .context.deploymentRegion ${SHELL_PATH}/../cdk.json)"
    if [ -z "$DEPLOYMENT_REGION" ]; then
        DEPLOYMENT_REGION="$(aws configure get region)"
    fi

    aws iam create-role \
        --role-name "${ROLE_NAME}" \
        --assume-role-policy-document '{
                                            "Version": "2012-10-17",
                                            "Statement": [
                                                {
                                                    "Effect": "Allow",
                                                    "Principal": {
                                                        "Service": "grafana.amazonaws.com"
                                                    },
                                                    "Action": "sts:AssumeRole"
                                                }
                                            ]
                                        }'

    local WORKSPACE_ID="$(aws grafana create-workspace \
                            --workspace-name ${WORKSPACE_NAME} \
                            --stack-set-name ${WORKSPACE_NAME} \
                            --account-access-type CURRENT_ACCOUNT \
                            --authentication-providers AWS_SSO \
                            --permission-type SERVICE_MANAGED \
                            --workspace-role-arn ${ROLE_NAME} \
                            --region ${DEPLOYMENT_REGION} \
                            --output text \
                            --query workspace.id)"

    local WORKSPACE_STATUS="$(aws grafana describe-workspace \
                            --workspace-id ${WORKSPACE_ID} \
                            --region ${DEPLOYMENT_REGION} \
                            --output text \
                            --query workspace.status)"

    while [ "$WORKSPACE_STATUS" != "ACTIVE" ]; do
        echo "Please Wait...${WORKSPACE_STATUS}"
        sleep 10
        WORKSPACE_STATUS="$(aws grafana describe-workspace \
                            --workspace-id ${WORKSPACE_ID} \
                            --region ${DEPLOYMENT_REGION} \
                            --output text \
                            --query workspace.status)"
    done

    local WORKSPACE_ENDPOINT="$(aws grafana describe-workspace \
                            --workspace-id ${WORKSPACE_ID} \
                            --region ${DEPLOYMENT_REGION} \
                            --output text \
                            --query workspace.endpoint)"

    jq -n --arg wsName ${WORKSPACE_NAME} \
        --arg wsId ${WORKSPACE_ID} \
        --arg wsEndpoint ${WORKSPACE_ENDPOINT} \
        --arg wsRole ${ROLE_NAME} \
        --arg region ${DEPLOYMENT_REGION} \
        '{
                "workspaceName": $wsName,
                "workspaceEndpoint": $wsEndpoint,
                "workspaceId":$wsId,
                "workspaceRole":$wsRole,
                "deploymentRegion":$region
        }' > ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE}

    echo "Done."
}

deleteWorkspace()
{
    local WORKSPACE_ID="$(jq -r .workspaceId ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE})"
    local ROLE_NAME="$(jq -r .workspaceRole ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE})"
    local DEPLOYMENT_REGION="$(jq -r .deploymentRegion ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE})"

    aws grafana delete-workspace --workspace-id ${WORKSPACE_ID} --region ${DEPLOYMENT_REGION}

    local WORKSPACE_STATUS="$(aws grafana describe-workspace \
                            --workspace-id ${WORKSPACE_ID} \
                            --region ${DEPLOYMENT_REGION} \
                            --output text \
                            --query workspace.status)"

    while [ "$WORKSPACE_STATUS" == "DELETING" ]; do
        echo "Please Wait...${WORKSPACE_STATUS}"
        sleep 10
        WORKSPACE_STATUS="$(aws grafana describe-workspace \
                            --workspace-id ${WORKSPACE_ID} \
                            --region ${DEPLOYMENT_REGION} \
                            --output text \
                            --query workspace.status \
                            2>/dev/null)"
    done

    aws iam delete-role --role-name "${ROLE_NAME}"
    rm -rf ${SHELL_PATH}/${GRAFANA_WORKSPACE_FILE}
    rm -rf ${SHELL_PATH}/${GRAFANA_DATASOURCE_FILE}

    echo "Done."
}

if [ $script_action = create ]; then
    createWorkspace
else
    deleteWorkspace
fi
