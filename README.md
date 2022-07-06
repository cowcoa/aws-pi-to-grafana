## AWS PI to Grafana
Sync AWS RDS/Aurora's Performance Insights metrics to Managed Grafana.<br />

## Prerequisites
1. Install and configure AWS CLI Version 2 environment:<br />
   [Installation] - Installing or updating the latest version of the AWS CLI v2.<br />
   [Configuration] - Configure basic settings that AWS CLI uses to interact with AWS.<br />
   NOTE: Make sure your IAM User/Role has sufficient permissions.
2. Install Node Version Manager:<br />
   [Install NVM] - Install NVM and configure your environment according to this document.
3. Install Node.js:<br />
    ```sh
    nvm install 16.3.0
    ```
4. Install AWS CDK Toolkit:
    ```sh
    npm install -g aws-cdk
    ```
5. Install Golang:<br />
   [Download and Install] - Download and install Go quickly with the steps described here.
6. Install Docker:<br />
   [Install Docker Engine] - The installation section shows you how to install Docker on a variety of platforms.
7. Make sure you also have GNU Make, jq installed:<br />
    ```sh
    sudo yum install -y make
    sudo yum install -y jq
    ```
## Deployment
Run the following command to deploy AWS infra and code by CDK Toolkit:<br />
  ```sh
  cdk-cli-wrapper-dev.sh deploy
  ```
You can also clean up the deployment by running command:<br />
  ```sh
  cdk-cli-wrapper-dev.sh destroy
  ```
Run the following command to create AWS Managed Grafana workspace:<br />
  ```sh
  grafana/create-grafana-workspace.sh create
  ```
After the above command is executed, you can create SSO user through AWS console and login to the Grafana workspace.
This command will also save the Grafana workspace information to grafana/grafana-workspace-info.json file.<br />
You can also clean up the Grafana workspace by running command:<br />
  ```sh
  grafana/create-grafana-workspace.sh delete
  ```
Run the following command to create default MySQL DataSource in Grafana workspace:<br />
  ```sh
  grafana/create-grafana-datasource.sh
  ```
This command will also save the MySQL database information to grafana/mysql-datasource-info.json file.

## Examples

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Install NVM]: <https://github.com/nvm-sh/nvm#install--update-script>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
