### Cloud Technologies (Spring 2020) - Bitly Clone

By Ward Huang

##### Deployment Diagram
<details>
<summary>screenshot</summary>

![Bitly-Clone Deployment Diagram 1](design/01.bitly-diagram.png)

</details>

##### Project networks ( with GKE cluster)
<details>
<summary>screenshots</summary>

[VPC and subnetworks](design/02.bitly-network.png)

[Firewall Rules](design/03.bitky-network-firewall.png)

[Instances and Network Tags](design/04.bitly-network-tags-and-instances.png)

</details>

##### Google Kubernetes Engine
<details>
<summary>screenshots</summary>

[Cluster Info](design/05.bitly-gke.png)

[Service Endpoints](design/06.bitly-gke-service-endpoints.png)

[Load Balancer](design/07.bitly-gke-load-balancer.png)

[Testing Service NodePort](design/08.bitly-gke-node-ports.png)

</details>

##### Testing API through Kong
<details>
<summary>screenshots</summary>

[Declarative Kong Config](design/09.bitly-kong-declarative-config.png)

[Create some Shortlinks](design/10.bitly-create-shortlinks.png)

[Visit some URLs](design/11.bitly-visit-shortlinks.png)

[Check trending URLs](design/12.bitly-trending-links.png)

</details>

##### CQRS
<details>
<summary>CQRS</summary>

Reads and Writes are handled differently in this system.
- Writes are passed onto the message bus.
- Reads have direct access to data storage.

[CQRS](design/13.bitly-cqrs.jpeg)

[CQRS-2](design/14.bitly-cqrs-2.jpeg)

</details>

##### Event Sourcing
<details>
<summary>Event Sourcing</summary>

Two event types are described below. These events are logged in the event-store.
- cp.shortlink.create --- create new shortlink
- lr.shortlink.update --- a user has visited a shortlink

[Event-Sourcing](design/15.bitly-event-sourcing.png)
</details>

##### Some issues faced during project ( Journal )
<details>
<summary>Project Journal</summary>

- **Problem**: how to access rabbitmq dashboard over internet
  - _Solution_: SSH Tunnel through google cloud SDK
- **Problem**: Instances crash on start up because they cannot connect to their databases
  - _Solution-1_: Open all ports to addresses in VPC CIDR block
  - _Solution-2_: Separate instances into logical subnetworks and configure firewall rules using network tags
- **Problem**: Changing environment variables in Dockerfile requires re-building docker image.
  - _Solution_: Set environment variables on Compute Engine or GKE deployment start-up
- **Problem**: GKE fails to pull image from docker hub
  - _Solution_: Push to Google's container registry GCR
- **Problem**: Container images are too large to push on DSL connection.
  - _Solution-1_: Use alpine linux for golang containers
  - _Solution-2_: Remove yum install package command from NoSQL Dockerfile
- **Problem**: Where do we get new shortlinks from?
  - _Solution-1_: Use inserted MySQL index ( increment a counter )
  - _Solution-2_: Hash the URL
  - _Solution-3_: Randomize strings and maintain key-value table of assigned shortlinks
  - _Solution-4_: Pre-generate shortlinks and allocate as needed

</details>

##### Control Panel API
<details>
<summary>Control Panel API</summary>

**/cp/ping --- GET**
```
Sample Response:
{
  "CP Server: a957fa7d-e579-494c-b25e-b822afd500e2 - API version 3.0 alive!"
}
```
**/cp/link_save --- POST**
```
Sample Request:
{
  "OrigUrl":"aws.amazon.com"
}
```
```
Sample Response:
{
  "ShortUrl": "vA"
}
```

</details>

##### Link Redirect API
<details>
<summary>Link Redirect API</summary>

**/lr/ping --- GET**
```
Sample Response:
{
  "LR Server: 2df9c7bd-04d8-4427-a193-29d77594e3ff - API version 3.0 alive!"
}
```
**/lr/r/{shortlink} --- GET**
```
Sample Response:
{
  "OrigUrl": "aws.amazon.com"
}
```

</details>

##### Trend Server API
<details>
<summary>Trend Server API</summary>

**/ts/ping --- GET**
```
Sample Response:
{
  "Test": "TS Server: 5f75372c-89ff-48e6-ba79-359bef38f546 - API version 3.0 alive!"
}
```
**/ts/t/{shortlink} --- GET**
```
Sample Response:
{
  "OrigUrl": "aws.amazon.com",
  "ShortUrl": "vA",
  "Visits": 2
}
```
**/ts/t/merge --- GET**
```
Sample Response:
[
  {
    "OrigUrl": "ifconfig.co",
    "Visits": 3
  },
  {
    "OrigUrl": "aws.amazon.com",
    "Visits": 2
  },
  {
    "OrigUrl": "gobyexample.com",
    "Visits": 2
  },
  {
    "OrigUrl": "one.sjsu.edu",
    "Visits": 1
  },
  {
    "OrigUrl": "sjsu.instructure.com",
    "Visits": 1
  }
]
```
</details>

#### Database Schemas
<details>
<summary>Database Schemas</summary>

**Main MySQL Database ( tiny_urls )**

key|name     |type
---|---------|-----------
PK |id       |bigint(20)
&nbsp;|orig_url |varchar(512)
&nbsp;|short_url|varchar(45)
&nbsp;|visits   |int(11)
&nbsp;|created  |timestamp

**Event-Store ( eventlogs )**

key|name      |type
---|----------|-----------
PK |\_id      |TimeUUID  
&nbsp;|routingkey|String  
&nbsp;|body      |QueueMessage

**UDT: QueueMessage**
```
{
  "origurl" : "cloud.google.com",
  "shorturl" : "A4"
}
```

**Trend Server ( visits )**

key|name           |type
---|---------------|-----------
PK |\_id (shorturl)|String  
&nbsp;|origurl        |String  
&nbsp;|visits         |Integer  

**Control Panel ( shortlinks )**

key|name           |type
---|---------------|-----------
PK |\_id (shorturl)|String  

**LR Cache ( NoSQL Project )**

key|name           |type
---|---------------|-----------
PK |key (shorturl) |String  
&nbsp;|value (origurl)|String  
</details>
---
