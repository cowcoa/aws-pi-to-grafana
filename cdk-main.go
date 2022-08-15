package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	secretmgr "github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"rds-mysql-cluster/config"
)

type RdsMySqlClusterStackProps struct {
	awscdk.StackProps
}

func NewRdsMySqlClusterStack(scope constructs.Construct, id string, props *RdsMySqlClusterStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Import default VPC.
	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("DefaultVPC"), &awsec2.VpcLookupOptions{
		IsDefault: jsii.Bool(true),
	})
	// Create MySQL 3306 inbound Security Group.
	sg := awsec2.NewSecurityGroup(stack, jsii.String("MySQLSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(*stack.StackName() + "-MySQLSG"),
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("RDS MySQL DB instances communication SG."),
	})
	sg.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			FromPort:             jsii.Number(3306),
			ToPort:               jsii.Number(3306),
			StringRepresentation: jsii.String("Standard MySQL listen port."),
		}),
		jsii.String("Allow requests to MySQL DB instance."),
		jsii.Bool(false),
	)
	// Database engine version.
	engine := awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
		Version: awsrds.MysqlEngineVersion_VER_5_7_34(),
	})
	// Database subnet group.
	subnetGrp := awsrds.NewSubnetGroup(stack, jsii.String("SubnetGroup"), &awsrds.SubnetGroupProps{
		Vpc:             vpc,
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
		SubnetGroupName: jsii.String(*stack.StackName() + "-SubnetGroup"),
		VpcSubnets:      &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
		Description:     jsii.String("Custom SubnetGroup"),
	})
	// Database parameter group.
	// https://aws.amazon.com/blogs/database/best-practices-for-configuring-parameters-for-amazon-rds-for-mysql-part-1-parameters-related-to-performance/
	paramGrp := awsrds.NewParameterGroup(stack, jsii.String("ParameterGroup"), &awsrds.ParameterGroupProps{
		Engine:      engine,
		Description: jsii.String("Custom ParameterGroup"),
		Parameters: &map[string]*string{
			"event_scheduler":        jsii.String("ON"),
			"innodb_sync_array_size": jsii.String("16"),
		},
	})
	// Database credential in SecretManager
	dbSecret := secretmgr.NewSecret(stack, jsii.String("DBSecret"), &secretmgr.SecretProps{
		SecretName: jsii.String(*stack.StackName() + "-Secret"),
		GenerateSecretString: &secretmgr.SecretStringGenerator{
			SecretStringTemplate: jsii.String(string(`{"username":"cow"}`)),
			ExcludePunctuation:   jsii.Bool(true),
			IncludeSpace:         jsii.Bool(false),
			GenerateStringKey:    jsii.String("password"),
		},
	})
	// Create RDS MySQL DB instance.
	dbPrimInstance := awsrds.NewDatabaseInstance(stack, jsii.String("PrimaryDBInstance"), &awsrds.DatabaseInstanceProps{
		Vpc:                     vpc,
		AutoMinorVersionUpgrade: jsii.Bool(true),
		BackupRetention:         awscdk.Duration_Days(jsii.Number(7)),
		CloudwatchLogsExports: &[]*string{
			jsii.String("error"),
			jsii.String("general"),
			jsii.String("slowquery"),
		},
		CloudwatchLogsRetention:     awslogs.RetentionDays_FIVE_DAYS,
		CopyTagsToSnapshot:          jsii.Bool(true),
		DeleteAutomatedBackups:      jsii.Bool(true),
		DeletionProtection:          jsii.Bool(false),
		EnablePerformanceInsights:   jsii.Bool(true),
		IamAuthentication:           jsii.Bool(false),
		InstanceIdentifier:          jsii.String(*stack.StackName() + "-PrimaryDBInstance"),
		Iops:                        jsii.Number(2000),
		MaxAllocatedStorage:         jsii.Number(100),
		MonitoringInterval:          awscdk.Duration_Seconds(jsii.Number(60)),
		MultiAz:                     jsii.Bool(true),
		ParameterGroup:              paramGrp,
		PerformanceInsightRetention: awsrds.PerformanceInsightRetention_DEFAULT,
		Port:                        jsii.Number(3306),
		PreferredBackupWindow:       jsii.String("15:30-16:30"),
		PreferredMaintenanceWindow:  jsii.String("wed:16:40-wed:17:40"),
		PubliclyAccessible:          jsii.Bool(true),
		RemovalPolicy:               awscdk.RemovalPolicy_DESTROY,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			sg,
		},
		StorageType:              awsrds.StorageType_GP2,
		SubnetGroup:              subnetGrp,
		Engine:                   engine,
		AllocatedStorage:         jsii.Number(20),
		AllowMajorVersionUpgrade: jsii.Bool(false),
		DatabaseName:             jsii.String(config.MySqlConnection.Database),
		InstanceType:             awsec2.InstanceType_Of(awsec2.InstanceClass_MEMORY5, awsec2.InstanceSize_LARGE),
		Credentials:              awsrds.Credentials_FromSecret(dbSecret, jsii.String(config.MySqlConnection.User)),
		StorageEncrypted:         jsii.Bool(false),
	})

	awsrds.NewDatabaseInstanceReadReplica(stack, jsii.String("ReplicaDBInstance"), &awsrds.DatabaseInstanceReadReplicaProps{
		InstanceIdentifier: jsii.String("ReplicaDBInstance"),
		Vpc:                vpc,
		ParameterGroup:     paramGrp,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			sg,
		},
		SubnetGroup:            subnetGrp,
		InstanceType:           awsec2.InstanceType_Of(awsec2.InstanceClass_MEMORY5, awsec2.InstanceSize_LARGE),
		SourceDatabaseInstance: dbPrimInstance,
		StorageEncrypted:       jsii.Bool(false),
	})

	config.MySqlConnection.Host = *dbPrimInstance.InstanceEndpoint().Hostname()

	// Output data source info.
	awscdk.NewCfnOutput(stack, jsii.String("host"), &awscdk.CfnOutputProps{
		Value: jsii.String(config.MySqlConnection.Host),
	})
	awscdk.NewCfnOutput(stack, jsii.String("database"), &awscdk.CfnOutputProps{
		Value: jsii.String(config.MySqlConnection.Database),
	})
	awscdk.NewCfnOutput(stack, jsii.String("user"), &awscdk.CfnOutputProps{
		Value: jsii.String(config.MySqlConnection.User),
	})
	awscdk.NewCfnOutput(stack, jsii.String("password"), &awscdk.CfnOutputProps{
		Value: jsii.String(config.MySqlConnection.Password),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewRdsMySqlClusterStack(app, config.StackName(app), &RdsMySqlClusterStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEPLOY_ACCOUNT")
	region := os.Getenv("CDK_DEPLOY_REGION")

	if len(account) == 0 || len(region) == 0 {
		account = os.Getenv("CDK_DEFAULT_ACCOUNT")
		region = os.Getenv("CDK_DEFAULT_REGION")
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
