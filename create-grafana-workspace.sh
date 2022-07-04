#!/bin/bash

WORKSPACE_NAME="zxaws-grafana-poc-3"
ROLE_NAME="${WORKSPACE_NAME}-role"
SSO_ID="9567711734-ada7d82f-ce20-4061-a509-ee75aa76789a"
DEPLOYMENT_REGION="ap-northeast-1"

aws iam create-role \
    --role-name "${ROLE_NAME}" \
    --assume-role-policy-document file://role-trust-policy.json \
    --region ${DEPLOYMENT_REGION}

aws iam create-role \
    --role-name "${ROLE_NAME}" \ 
    --region "${DEPLOYMENT_REGION}" \
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

WORKSPACE_ID="$(aws grafana create-workspace \
                    --workspace-name ${WORKSPACE_NAME} \
                    --stack-set-name ${WORKSPACE_NAME} \
                    --account-access-type CURRENT_ACCOUNT \
                    --authentication-providers AWS_SSO \
                    --permission-type SERVICE_MANAGED \
                    --workspace-role-arn ${ROLE_NAME} \
                    --region ${DEPLOYMENT_REGION} \
                    --output text \
                    --query workspace.id)"

echo "WORKSPACE_ID:${WORKSPACE_ID}"

WORKSPACE_STATUS="$(aws grafana describe-workspace \
                        --workspace-id ${WORKSPACE_ID} \
                        --region ${DEPLOYMENT_REGION} \
                        --output text \
                        --query workspace.status)"

while [ "$WORKSPACE_STATUS" != "ACTIVE" ]; do
	echo "Please Wait..."
	sleep 10
    WORKSPACE_STATUS="$(aws grafana describe-workspace \
                        --workspace-id ${WORKSPACE_ID} \
                        --region ${DEPLOYMENT_REGION} \
                        --output text \
                        --query workspace.status)"
done

aws grafana update-permissions \
        --workspace-id ${WORKSPACE_ID} \
        --update-instruction-batch action="ADD",role="ADMIN",users=["{id=${SSO_ID},type=SSO_USER}"] \
        --region ${DEPLOYMENT_REGION}
