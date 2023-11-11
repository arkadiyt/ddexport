package ddexport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type DDExport struct {
	counter int
	query   string
	to      string
	from    string
	limit   int
	output  io.Writer
}

func New(query string, to string, from string, limit int, output io.Writer) *DDExport {
	d := DDExport{}
	d.query = query
	d.to = to
	d.from = from
	d.limit = limit
	d.output = output
	return &d
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

func (d *DDExport) client() (*datadog.APIClient, context.Context) {
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: os.Getenv("DD_API_KEY"),
			},
			"appKeyAuth": {
				Key: os.Getenv("DD_APP_KEY"),
			},
		},
	)

	configuration := datadog.NewConfiguration()
	configuration.RetryConfiguration.EnableRetry = true
	configuration.RetryConfiguration.MaxRetries = 10

	apiClient := datadog.NewAPIClient(configuration)
	return apiClient, ctx
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
	client, ctx := d.client()
	api := datadogV2.NewLogsApi(client)

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

	resp, _ := api.ListLogsWithPagination(ctx, *datadogV2.NewListLogsOptionalParameters().WithBody(body))
	writeRecords(d, resp)
}

func (d *DDExport) SearchSpans() {
	client, ctx := d.client()
	api := datadogV2.NewSpansApi(client)

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

	resp, _ := api.ListSpansWithPagination(ctx, body)
	writeRecords(d, resp)
}
