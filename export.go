package ddexport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/sethvargo/go-envconfig"
)

type DDExport struct {
	counter    int
	query      string
	to         string
	from       string
	limit      int
	output     io.Writer
	ApiKeyAuth string `env:"DD_API_KEY,required"`
	AppKeyAuth string `env:"DD_APP_KEY,required"`
	apiClient  *datadog.APIClient
	ctx        context.Context
}

func New(query, to, from string, limit int, output io.Writer) (*DDExport, error) {
	d := DDExport{}
	d.query = query
	d.to = to
	d.from = from
	d.limit = limit
	d.output = output

	err := envconfig.Process(context.Background(), &d)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: d.ApiKeyAuth,
			},
			"appKeyAuth": {
				Key: d.AppKeyAuth,
			},
		},
	)

	configuration := datadog.NewConfiguration()
	configuration.RetryConfiguration.EnableRetry = true
	configuration.RetryConfiguration.MaxRetries = 10
	apiClient := datadog.NewAPIClient(configuration)
	d.apiClient = apiClient
	d.ctx = ctx

	return &d, nil
}

func writeRecords[K any](d *DDExport, resp <-chan datadog.PaginationResult[K]) {
	for paginationResult := range resp {
		if paginationResult.Error != nil {
			fmt.Fprintf(os.Stderr, "Api error: %v\n", paginationResult.Error)
		}

		json, _ := json.Marshal(paginationResult.Item)
		d.writeRecord(json)
	}
	d.progress()
}

func (d *DDExport) writeRecord(record []byte) {
	_, err := d.output.Write(record)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	d.output.Write([]byte("\n"))

	d.counter += 1
	if d.counter%1000 == 0 {
		d.progress()
	}
}

func (d *DDExport) progress() {
	fmt.Printf("Wrote %d records\n", d.counter)
}

func (d *DDExport) SearchLogs() {
	api := datadogV2.NewLogsApi(d.apiClient)

	body := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			From:  datadog.PtrString(d.from),
			To:    datadog.PtrString(d.to),
			Query: datadog.PtrString(d.query),
		},
		Options: &datadogV2.LogsQueryOptions{
			Timezone: datadog.PtrString("GMT"),
		},
		Page: &datadogV2.LogsListRequestPage{
			Limit: datadog.PtrInt32(int32(d.limit)),
		},
	}

	resp, _ := api.ListLogsWithPagination(d.ctx, *datadogV2.NewListLogsOptionalParameters().WithBody(body))
	writeRecords(d, resp)
}

func (d *DDExport) SearchSpans() {
	api := datadogV2.NewSpansApi(d.apiClient)

	body := datadogV2.SpansListRequest{
		Data: &datadogV2.SpansListRequestData{
			Attributes: &datadogV2.SpansListRequestAttributes{
				Filter: &datadogV2.SpansQueryFilter{
					From:  datadog.PtrString(d.from),
					To:    datadog.PtrString(d.to),
					Query: datadog.PtrString(d.query),
				},
				Options: &datadogV2.SpansQueryOptions{
					Timezone: datadog.PtrString("GMT"),
				},
				Page: &datadogV2.SpansListRequestPage{
					Limit: datadog.PtrInt32(int32(d.limit)),
				},
			},
			Type: datadogV2.SPANSLISTREQUESTTYPE_SEARCH_REQUEST.Ptr(),
		},
	}

	resp, _ := api.ListSpansWithPagination(d.ctx, body)
	writeRecords(d, resp)
}
