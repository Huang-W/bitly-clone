# bit.ly clone
### Functional Requirements

### Public API

### RabbitMQ Channels

- cp.shortlink.create
- lr.shortlink.read (RPC)
- lr.shortlink.update

### Database Schema

key|name     |type
---|---------|-----------
PK |id       |bigint(20)
   |orig_url |varchar(512)
   |short_url|varchar(45)
   |visits   |int(11)
   |created  |timestamp
