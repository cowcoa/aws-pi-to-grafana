package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"pi-to-grafana/config"
)

type PiToGrafanaStackProps struct {
	awscdk.StackProps
}

func NewPiToGrafanaStack(scope constructs.Construct, id string, props *PiToGrafanaStackProps) awscdk.Stack {
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

	// Create RDS MySQL DB instance.
	dbInstance := awsrds.NewDatabaseInstance(stack, jsii.String("GrafanaDataSource"), &awsrds.DatabaseInstanceProps{
		InstanceIdentifier: jsii.String(*stack.StackName() + "-GrafanaDataSource"),
		Vpc:                vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			sg,
		},
		Engine: awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
			Version: awsrds.MysqlEngineVersion_VER_8_0_26(),
		}),
		DatabaseName:              jsii.String(config.MySqlConnection.Database),
		InstanceType:              awsec2.InstanceType_Of(awsec2.InstanceClass_STANDARD5, awsec2.InstanceSize_LARGE),
		StorageType:               awsrds.StorageType_GP2,
		AllocatedStorage:          jsii.Number(20),
		MaxAllocatedStorage:       jsii.Number(100),
		Credentials:               awsrds.Credentials_FromPassword(jsii.String(config.MySqlConnection.User), awscdk.SecretValue_PlainText(jsii.String(config.MySqlConnection.Password))),
		MultiAz:                   jsii.Bool(false),
		PubliclyAccessible:        jsii.Bool(true),
		EnablePerformanceInsights: jsii.Bool(true),
		AutoMinorVersionUpgrade:   jsii.Bool(false),
		CopyTagsToSnapshot:        jsii.Bool(false),
		DeleteAutomatedBackups:    jsii.Bool(true),
		DeletionProtection:        jsii.Bool(false),
		StorageEncrypted:          jsii.Bool(false),
	})

	config.MySqlConnection.Host = *dbInstance.InstanceEndpoint().Hostname()

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

	// Create role for lambda function.
	lambdaRole := awsiam.NewRole(stack, jsii.String("LambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(*stack.StackName() + "-LambdaRole"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchFullAccess")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonRDSPerformanceInsightsReadOnly")),
		},
	})

	// Create EventBridge trigger function.
	triggerFunction := awslambda.NewFunction(stack, jsii.String("DataCollector"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-DataCollector"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(256),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/data-collector/."), nil),
		Handler:      jsii.String("data-collector"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"MYSQL_HOST":         jsii.String(config.MySqlConnection.Host),
			"MYSQL_DATABASE":     jsii.String(config.MySqlConnection.Database),
			"MYSQL_USER":         jsii.String(config.MySqlConnection.User),
			"MYSQL_PASSWORD":     jsii.String(config.MySqlConnection.Password),
			"TARGET_INSTANCE_ID": jsii.String(config.TargetInstanceId(stack)),
		},
	})
	triggerFunction.Node().AddDependency(dbInstance)

	// Create EventBridge rule.
	awsevents.NewRule(stack, jsii.String("EventTrigger"), &awsevents.RuleProps{
		RuleName: jsii.String(*stack.StackName() + "-EventTrigger"),
		Enabled:  jsii.Bool(true),
		Schedule: awsevents.Schedule_Rate(awscdk.Duration_Minutes(jsii.Number(config.ScheduleRate))),
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewLambdaFunction(triggerFunction, &awseventstargets.LambdaFunctionProps{
				Event: awsevents.RuleTargetInput_FromText(jsii.String("Hello World!")),
			}),
		},
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewPiToGrafanaStack(app, config.StackName(app), &PiToGrafanaStackProps{
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
