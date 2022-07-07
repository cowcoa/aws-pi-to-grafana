## AWS PI to Grafana
Demonstrate how to sync AWS RDS/Aurora's Performance Insights metrics to Managed Grafana.<br />
Since Performance Insights only measures one DB instance at a time, so does this example, which can only sync PI metrics for a specified DB instance to the Grafana dashboard. But you can extend this example to the cluster level by yourself.

## Supported Metrics 

grafana.db_status:

| Metric | Unit | Description |
| ------ | ------ | ------ |
| db.SQL.Com_analyze | Queries per second | Number of ANALYZE commands executed |
| db.SQL.Com_optimize | Queries per second | Number of OPTIMIZE commands executed |
| db.SQL.Com_select | Queries per second | Number of SELECT commands executed |
| db.SQL.Innodb_rows_inserted | Rows per second | Total rows inserted by InnoDB |
| db.SQL.Innodb_rows_deleted | Rows per second | Total rows deleted by InnoDB |
| db.SQL.Innodb_rows_updated | Rows per second | Total rows updated by InnoDB |
| db.SQL.Innodb_rows_read | Rows per second | Total rows read by InnoDB |
| db.SQL.Questions | Queries per second | The number of statements executed by the server. This includes only statements sent to the server by clients and not statements executed within stored programs |
| db.SQL.Queries | Queries per second | The number of statements executed by the server. This variable includes statements executed within stored programs |
| db.SQL.Select_full_join | Queries per second | The number of joins that perform table scans because they do not use indexes. If this value is not 0 you should carefully check the indexes of your tables |
| db.SQL.Select_full_range_join | Queries per second | The number of joins that used a range search on a reference table |
| db.SQL.Select_range | Queries per second | The number of joins that used ranges on the first table. This is normally not a critical issue even if the value is quite large |
| db.SQL.Select_range_check | Queries per second | The number of joins without keys that check for key usage after each row. If this is not 0 you should carefully check the indexes of your tables |
| db.SQL.Select_scan | Queries per second | The number of joins that did a full scan of the first table |
| db.SQL.Slow_queries | Queries per second | The number of queries that have taken more than long_query_time seconds. This counter increments regardless of whether the slow query log is enabled |
| db.SQL.Sort_merge_passes | Queries per second | The number of merge passes that the sort algorithm has had to do. If this value is large you should consider increasing the value of the sort_buffer_size system variable |
| db.SQL.Sort_range | Queries per second | The number of sorts that were done using ranges |
| db.SQL.Sort_rows | Queries per second | The number of sorted rows |
| db.SQL.Sort_scan | Queries per second | The number of sorts that were done by scanning the table |
| db.Locks.Innodb_row_lock_time | Milliseconds | The total time spent in acquiring row locks for InnoDB tables in milliseconds. |
| db.Locks.innodb_row_lock_waits | Transactions | The number of times operations on InnoDB tables had to wait for a row lock |
| db.Locks.innodb_deadlocks | Deadlocks per minute | Number of deadlocks |
| db.Locks.innodb_lock_timeouts | Timeouts | Number of InnoDB lock timeouts |

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

## Configuration

You can edit the cdk.json file to modify the deployment configuration.

| Key | Example Value | Description |
| ------ | ------ | ------ |
| stackName | PI2Grafana | CloudFormation stack name. It's difficult for jq to process JSON keys containing '-', so avoid naming your stack that way. |
| deploymentRegion | ap-northeast-1 | CloudFormation stack deployment region. If the value is empty, the default is the same as the region where deploy is executed. |
| dbInstanceName | my-rds-db-instance | RDS/Aurora DB instance name. This instance is your monitoring target, so make sure it already exists before deploying this example. |

## Deployment
1. Run the following command to deploy AWS infra and code by CDK Toolkit:<br />
     ```sh
     cdk-cli-wrapper-dev.sh deploy
     ```
   You can also clean up the deployment by running command:<br />
     ```sh
     cdk-cli-wrapper-dev.sh destroy
     ```
2. Run the following command to create AWS Managed Grafana workspace:<br />
     ```sh
     grafana/create-grafana-workspace.sh create
     ```
   This command will also save the Grafana workspace information to grafana/grafana-workspace-info.json file.<br />
   You can also clean up the Grafana workspace by running command:<br />
     ```sh
     grafana/create-grafana-workspace.sh delete
     ```
3. Run the following command to create default MySQL DataSource in Grafana workspace:<br />
     ```sh
     grafana/create-grafana-datasource.sh
     ```
   This command will also save the MySQL database information to grafana/mysql-datasource-info.json file.
4. Sign in to your AWS Web Console.
5. You can refer to [this blog](https://aws.amazon.com/blogs/security/how-to-create-and-manage-users-within-aws-sso/) to create AWS SSO User.
6. Find the Grafana workspace you just created and [add the AWS SSO User](https://docs.aws.amazon.com/grafana/latest/userguide/AMG-manage-users-and-groups-AMG.html) as the ADMIN.
7. Sign in to the Grafana workspace and find the MySQL DataSource you just created, fill in the password(according to the grafana/mysql-datasource-info.json file), and click "Save & Test".
8. Import Grafana dashboard by upload grafana/grafana-dashboard.json file.

## Examples

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Install NVM]: <https://github.com/nvm-sh/nvm#install--update-script>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
