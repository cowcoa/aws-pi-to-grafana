package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// MySQL DataSource info.
type DataSource struct {
	Host     string
	Database string
	User     string
	Password string
}

// DO NOT keep DB info here.
// This is just for convenience of testing.
var MySqlConnection = DataSource{
	Host:     "learn-grafana.c2jzihkjutmr.ap-northeast-1.rds.amazonaws.com:3306",
	Database: "my2",
	User:     "my2",
	Password: "1234abcd",
}

const (
	// EventBridge trigger Lambda function per minutes.
	ScheduleRate = 5
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "PI-to-Grafana"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// The DB instance identifier you want to monitor.
// DO NOT modify this function, change DB instance name by 'cdk.json/context/dbInstanceName'.
// Then the DB instance identifier will be automatically extracted by instance name:
// 'aws rds describe-db-instances --db-instance-identifier my-db-instance --region my-region --output text --query DBInstances[0].DbiResourceId'
func TargetInstanceId(scope constructs.Construct) string {
	dbInstanceId := "db-abcd1234"

	ctxValue := scope.Node().TryGetContext(jsii.String("dbInstanceId"))
	if v, ok := ctxValue.(string); ok {
		dbInstanceId = v
	}

	return dbInstanceId
}
