## AWS PI to Grafana.
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

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Install NVM]: <https://github.com/nvm-sh/nvm#install--update-script>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
