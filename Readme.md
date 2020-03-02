# Deployment Log Adapter

This adapter grabs logs for the `deploy.sh` scripts for adapter and publishes those logs on to the `<edge_name>/_platform` topic

## Setup

In order for this adapter to work, we need to have the following steps done before invoking the adapter. 
- Create a system on the ClearBlade platform
- Create a user, by default it is assigned an authenticated role
- Go to the `roles` page on the console & select the authenticated role
- Assign `publish` and `subscribe` permissions to the `<edge_name>/_platform` topic
- Create an adapter, name it `test` to test the working of this log adapter
- Add deploy.sh, start.sh and stop.sh files to the adapter
- Create a new deployment, and add the adapter to deployments by going to the `deployments` page

## Adapter Ops

### Run adapter as a script

```bash
go run main.go -file=<ABSOLUTE_FILE_PATH> -systemKey=<SYSTEM_KEY> -systemSecret=<SYSTEM_SECRET> -platformURL=<PLATFORM_URL> -email=<USER_EMAIL> -password=<PASSWORD> -messagingURL=<MESSAGING_URL>
```

Example: 

```bash
go run main.go -file=nohup.out -systemKey=8ca1e3e4abaf0cf8fba01 -systemSecret=8CA1E3E40BCAE6949F8535 -platformURL=https://dev.clearblade.com -email=logger@clearblade.com -password=aaaaaa -messagingURL=dev.clearblade.com:1883
```

### Build and execute

```bash

GOOS=${os:-linux} GOARCH=${arch:-amd64} GOARM=$arm go build
```


