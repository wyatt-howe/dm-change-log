# dm-change-log
change log services for all services related to data-mart

### Flow
- Whenever any relevant service alters it's data store it is the responsibility of that service to write a `model.ChangeEvent` to kafka for the workers to pick up
- n workers will consume from that topic as distinct worker groups
    - worker_api will post the change events to the api (currently backed by mariadb)
    - worker_s3 will batch and post parquet encoded change events to s3
- the `api` will serve as a CRUD API for the change events.

### Notes
- this is a mono-repo for several microservices
- as it stands the api will not allow mutations but that may be revisisted
- spec defined in the data-mart [TRD](https://magnite.atlassian.net/wiki/spaces/Tech/pages/141262930/Datamart+-+MVP)
- as it stands the test stub requires mariadb and kafka to be opperational on localhost default ports (with the schema loaded and the topic extant or kafka configured for auto create etc)

### TODO
- worker_s3 (service to consume from the kafka topic and post batched change events to s3 as parquet)
- docker compose setup for dependencies (kafka, mariadb etc)
- proper tests