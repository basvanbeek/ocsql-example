# ocsql-example

Example showing how to use [ocsql] the [OpenCensus] driver wrapper for database/sql.

## zipkin

This demo uses [Zipkin] as the tracing backend as it is very easy to set-up.
See the [Zipkin quickstart](https://zipkin.io/pages/quickstart)

## usage

A typical way to see the example in action:
```bash
# run Zipkin from official docker image
docker run -d -p 9411:9411 openzipkin/zipkin

# build the server application
go build cmd/svc/svc.go

# start the service
./svc
```

Open your favorite web browser and try these URL's
```
# valid
http://localhost:8080/user/1/items
http://localhost:8080/user/2/items

# not found
http://localhost:8080/user/100/items

# bad param
http://localhost:8080/user/INVALID_USER_ID/items

```

If all is well you should be able to see some generated spans at http://localhost:9411

[ocsql]:(https://github.com/opencensus-integrations/ocsql)
[zipkin]:(https://zipkin.io)
[opencensus]:(https://opencensus.io)
