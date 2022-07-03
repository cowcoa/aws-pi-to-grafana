package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/pi/types"

	runtime "github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

// MySQL query row date
type SqlItem struct {
	VariableName  string
	VariableValue string
	Timest        time.Time
}

const (
	// Time duration between start and end time, in seconds.
	QueryDuration = 300
	// The granularity, in seconds, of the data points returned from Performance Insights.
	// 1 (one second)
	// 60 (one minute)
	// 300 (five minutes)
	// 3600 (one hour)
	// 86400 (twenty-four hours)
	PeriodInSeconds = 60
	//
	MetricTypeOs = "os"
	MetricTypeDb = "db"
)

// aws pi get-resource-metrics --start-time and --end-time
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

type MetricsQueryInfo struct {
	TargetId          string
	StartTime         time.Time
	EndTime           time.Time
	Granularity       int32
	MetricType        string
	MetricQueriesList [][]types.MetricQuery
}

func handleRequest(ctx context.Context, req string) error {
	log.Printf("DetailType = %s\n", req)

	region := os.Getenv("AWS_REGION")
	// hostname:port format
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDb := os.Getenv("MYSQL_DATABASE")
	mysqlUsr := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PASSWORD")
	//
	targetId := os.Getenv("TARGET_INSTANCE_ID")

	log.Printf("AWS_REGION: %s.\n", region)
	log.Printf("MYSQL_HOST: %s.\n", mysqlHost)
	log.Printf("MYSQL_DATABASE: %s.\n", mysqlDb)
	log.Printf("MYSQL_USER: %s.\n", mysqlUsr)
	log.Printf("MYSQL_PASSWORD: %s.\n", mysqlPwd)
	log.Printf("TARGET_INSTANCE_ID: %s.\n", targetId)

	log.Println("Add code to connect with mysql server")
	dbClient, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", mysqlUsr, mysqlPwd, mysqlHost, mysqlDb))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer dbClient.Close()

	initStatusTables(dbClient)

	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Create PI client.
	piClient := pi.NewFromConfig(cfg)

	// Sync db metrics first.
	dbQueryInfo, err := constructQueryInfo(targetId, PeriodInSeconds, MetricTypeDb, dbClient, piClient, &ctx)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	syncMetrics(*dbQueryInfo, dbClient, piClient, &ctx)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	/*
		// Create table if not exists.
		sqlCreateTable := `
			CREATE TABLE IF NOT EXISTS status (
				VARIABLE_NAME varchar(64) CHARACTER SET utf8 NOT NULL DEFAULT '',
				VARIABLE_VALUE varchar(1024) CHARACTER SET utf8 DEFAULT NULL,
				TIMEST timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			) ENGINE=InnoDB;
		`
		_, err = db.Exec(sqlCreateTable)
		if err != nil {
			log.Println(err.Error())
			return err
		}
	*/

	// Sync os metrics first.
	/*
		results, err := db.Query("SELECT * FROM status ORDER BY timest DESC LIMIT 1")
		if err != nil {
			log.Println(err.Error())
			return err
		}

		var record SqlItem
		for results.Next() {
			err = results.Scan(&record.VariableName, &record.VariableValue, &record.Timest)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			log.Printf("query result: %+v\n", record)
		}

		var timeRange TimeRange
		if len(record.VariableName) > 0 {
			timeRange.StartTime = record.Timest
			timeRange.EndTime = timeRange.StartTime.Add(time.Second * time.Duration(QueryDuration))
		} else {
			timeRange.EndTime = time.Now()
			timeRange.StartTime = timeRange.EndTime.Add(-time.Second * time.Duration(QueryDuration))
		}
		log.Printf("query range: %+v\n", timeRange)

		// Load AWS configuration.
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			log.Println(err.Error())
			return err
		}

		piClient := pi.NewFromConfig(cfg)
		metricList, err := piClient.ListAvailableResourceMetrics(ctx, &pi.ListAvailableResourceMetricsInput{
			ServiceType: types.ServiceTypeRds,
			Identifier:  aws.String(target),
			MetricTypes: []string{
				"db",
			},
		})
		if err != nil {
			log.Println(err.Error())
			return err
		}

		var metricQueries []types.MetricQuery
		for i, metricName := range metricList.Metrics {
			log.Printf("metrics query name: %s\n", *metricName.Metric)
			//if *metricName.Metric == "db.SQL.Com_select" {
			if i < 15 {
				metricQueries = append(metricQueries, types.MetricQuery{Metric: aws.String(fmt.Sprintf("%s.avg", *metricName.Metric))})
			}
		}
		//log.Printf("metricQueries: %+v\n", metricQueries)
		for _, query := range metricQueries {
			log.Printf("query: %s\n", *query.Metric)
		}
	*/

	/*
		output, err := piClient.GetResourceMetrics(ctx, &pi.GetResourceMetricsInput{
			ServiceType:     types.ServiceTypeRds,
			Identifier:      aws.String(targetId),
			StartTime:       &timeRange.StartTime,
			EndTime:         &timeRange.EndTime,
			PeriodInSeconds: aws.Int32(PeriodInSeconds),

			MetricQueries: []types.MetricQuery{
				{
					Metric: aws.String("os.cpuUtilization.user.avg"),
				},
				{
					Metric: aws.String("db.SQL.Com_select.avg"),
				},
				{
					Metric: aws.String("db.SQL.Innodb_rows_inserted.avg"),
				},
			},
			//MetricQueries: metricQueries,
		})
		if err != nil {
			log.Println(err.Error())
			return err
		}

		// log.Printf("query result: %+v\n", output)

		for _, metric := range output.MetricList {
			log.Printf("metric key: %s\n", *metric.Key.Metric)

			for dname, dvalue := range metric.Key.Dimensions {
				log.Printf("Dimension: (%s, %s)\n", dname, dvalue)
			}

			for _, dataPoint := range metric.DataPoints {
				log.Printf("datapoint: (%s, %f)\n", dataPoint.Timestamp.String(), *dataPoint.Value)

				// Insert
				sqlInsertMetric := fmt.Sprintf(`INSERT INTO status(variable_name, variable_value, timest) VALUES ("%s", "%f", "%s")`, *metric.Key.Metric, *dataPoint.Value, dataPoint.Timestamp.UTC().Format("2006-01-02 15:04:05"))
				log.Printf("insert sql: %s\n", sqlInsertMetric)
				_, err = dbClient.Exec(sqlInsertMetric)
				if err != nil {
					log.Println(err.Error())
					return err
				}
			}
		}
	*/

	return nil
}

// Initalize metric status tables of MySQL data source db.
func initStatusTables(dbClient *sql.DB) error {
	// Create db status table if not exists.
	sqlCreateDbTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_status (
			VARIABLE_NAME varchar(64) CHARACTER SET utf8 NOT NULL DEFAULT '',
			VARIABLE_VALUE varchar(1024) CHARACTER SET utf8 DEFAULT NULL,
			TIMEST timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB;`, MetricTypeDb)
	_, err := dbClient.Exec(sqlCreateDbTable)
	if err != nil {
		return err
	}

	// Create os status table if not exists.
	sqlCreateOsTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_status (
			VARIABLE_NAME varchar(64) CHARACTER SET utf8 NOT NULL DEFAULT '',
			VARIABLE_VALUE varchar(1024) CHARACTER SET utf8 DEFAULT NULL,
			TIMEST timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB;`, MetricTypeOs)
	_, err = dbClient.Exec(sqlCreateOsTable)
	if err != nil {
		return err
	}

	return nil
}

func chunkSlice(slice []types.MetricQuery, chunkSize int) [][]types.MetricQuery {
	var chunks [][]types.MetricQuery
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// Construct PI metrics query info.
// Available metric type: db | os
func constructQueryInfo(queryTarget string, queryGranularity int32, metricType string,
	dbClient *sql.DB, piClient *pi.Client, ctx *context.Context) (*MetricsQueryInfo, error) {
	if metricType != MetricTypeDb && metricType != MetricTypeOs {
		return nil, fmt.Errorf("invalid metric type: %s", metricType)
	}

	var queryInfo *MetricsQueryInfo = new(MetricsQueryInfo)
	queryInfo.MetricType = metricType
	queryInfo.TargetId = queryTarget
	queryInfo.Granularity = queryGranularity

	// Try to get the recently item in db status table.
	sqlQuery := fmt.Sprintf(`SELECT * FROM %s_status ORDER BY timest DESC LIMIT 1`, metricType)
	sqlResults, err := dbClient.Query(sqlQuery)
	if err != nil {
		return nil, err
	}

	var item SqlItem
	for sqlResults.Next() {
		err = sqlResults.Scan(&item.VariableName, &item.VariableValue, &item.Timest)
		if err != nil {
			continue
		}
		log.Printf("query result: %+v\n", item)
	}

	// Compute query time range
	if len(item.VariableName) > 0 {
		queryInfo.StartTime = item.Timest
		queryInfo.EndTime = queryInfo.StartTime.Add(time.Second * time.Duration(QueryDuration))
	} else {
		queryInfo.EndTime = time.Now()
		queryInfo.StartTime = queryInfo.EndTime.Add(-time.Second * time.Duration(QueryDuration))
	}
	log.Printf("query range: (%s,%s)\n", queryInfo.StartTime.String(), queryInfo.EndTime.String())

	// get metric queries
	metricList, err := piClient.ListAvailableResourceMetrics(*ctx, &pi.ListAvailableResourceMetricsInput{
		ServiceType: types.ServiceTypeRds,
		Identifier:  aws.String(queryInfo.TargetId),
		MetricTypes: []string{
			queryInfo.MetricType,
		},
	})
	if err != nil {
		return nil, err
	}
	log.Printf("metrics returned: %d\n", len(metricList.Metrics))

	var metricQueries []types.MetricQuery
	for _, metricName := range metricList.Metrics {
		log.Printf("metrics query name: %s\n", *metricName.Metric)
		metricQueries = append(metricQueries, types.MetricQuery{Metric: aws.String(fmt.Sprintf("%s.avg", *metricName.Metric))})
	}
	/*
		for _, query := range metricQueries {
			log.Printf("query: %s\n", *query.Metric)
		}
	*/

	queryInfo.MetricQueriesList = chunkSlice(metricQueries, 10)

	/*
		for _, metricQueries := range queryInfo.MetricQueriesList {
			for _, query := range metricQueries {
				log.Printf("query: %s\n", *query.Metric)
			}
		}
	*/

	return queryInfo, nil
}

func syncMetrics(queryInfo MetricsQueryInfo, dbClient *sql.DB, piClient *pi.Client, ctx *context.Context) error {
	for _, metricQueries := range queryInfo.MetricQueriesList {
		log.Println("..... in syncMetrics .....")
		for _, query := range metricQueries {
			log.Printf("query: %s\n", *query.Metric)
		}

		metricsInfo, err := piClient.GetResourceMetrics(*ctx, &pi.GetResourceMetricsInput{
			ServiceType:     types.ServiceTypeRds,
			Identifier:      aws.String(queryInfo.TargetId),
			StartTime:       &queryInfo.StartTime,
			EndTime:         &queryInfo.EndTime,
			PeriodInSeconds: aws.Int32(queryInfo.Granularity),
			MetricQueries:   metricQueries,
		})
		if err != nil {
			return err
		}

		for _, metric := range metricsInfo.MetricList {
			log.Printf("metric key: %s\n", *metric.Key.Metric)

			for dname, dvalue := range metric.Key.Dimensions {
				log.Printf("Dimension: (%s, %s)\n", dname, dvalue)
			}

			for _, dataPoint := range metric.DataPoints {
				log.Printf("datapoint: (%s, %f)\n", dataPoint.Timestamp.String(), *dataPoint.Value)

				// Insert
				sqlInsertMetric := fmt.Sprintf(`INSERT INTO %s_status(variable_name, variable_value, timest) VALUES ("%s", "%f", "%s")`, queryInfo.MetricType, *metric.Key.Metric, *dataPoint.Value, dataPoint.Timestamp.UTC().Format("2006-01-02 15:04:05"))
				log.Printf("insert sql: %s\n", sqlInsertMetric)
				_, err = dbClient.Exec(sqlInsertMetric)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	runtime.Start(handleRequest)
}
