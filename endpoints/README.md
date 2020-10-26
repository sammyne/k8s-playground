# endpoints

Querying the endpoints named the same as a service can find out the endpoints/Pods backing the services.

@see https://kubernetes.io/docs/concepts/services-networking/connect-applications-service/#creating-a-service

## Quickstart

```bash
go run main.go --user {your-name} \
  --token {your-token}            \
  --master {your-api-server-url}  \
  --endpoint {endpoint-name}
```
