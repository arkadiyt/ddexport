# ddexport

Datadog's UI only lets you export 5000 rows of logs or spans - `ddexport` is a utility to hit Datadog's api and export the complete set of data.

## Usage

```
go install github.com/arkadiyt/ddexport/cmd@latest

# Set your API and APP keys
export DD_API_KEY=...
export DD_APP_KEY=...

# Export logs
ddexport logs -query 'env:production' -from 'now-30d' -to 'now' -output output.txt

# Export spans
ddexport spans -query 'env:production' -from 'now-30d' -to 'now' -output output.txt
```

## Getting in touch

Feel free to contact me on twitter: https://twitter.com/arkadiyt
