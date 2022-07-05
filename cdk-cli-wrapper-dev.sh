#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(jq -r .context.deploymentRegion ./cdk.json)"

# Check execution env.
if [ -z $CODEBUILD_BUILD_ID ]
then
    if [ -z "$CDK_REGION" ]; then
        CDK_REGION="$(aws configure get region)"
    fi

    echo "Run bootstrap..."
    export CDK_NEW_BOOTSTRAP=1 
    npx cdk bootstrap aws://${CDK_ACC}/${CDK_REGION} --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
else
    CDK_REGION=$AWS_DEFAULT_REGION
fi

# CDK command pre-process.
pushd ./functions &> /dev/null
    # Compile to x86_64 regardless of local arch.
    # This is the target arch that we will deploy to Lambda service.
    make TARGET_DIR="." GO_ARCH="amd64"
popd &> /dev/null

# CDK command.
DB_INSTANCE_NAME="$(jq -r .context.dbInstanceName ./cdk.json)"
DB_INSTANCE_IDENTIFIER="$(aws rds describe-db-instances \
                            --db-instance-identifier ${DB_INSTANCE_NAME} \
                            --region ${CDK_REGION} \
                            --output text \
                            --query DBInstances[0].DbiResourceId)"
set -- "$@" "-c" "dbInstanceId=${DB_INSTANCE_IDENTIFIER}" "--outputs-file" "${SHELL_PATH}/cdk.out/datasource-info.json"
$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"

# CDK command post-process.
if [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/
fi
