# bit.ly clone
### Functional Requirements
- Shorten a link
  - Input: a web link
  - Output: the shortened link
- Expand a link
  - Input: a shortened web link
  - Output: the expanded link
- Get trending links
  - Output: a list of links

### Public API
- /
  - GET - health check
- / urls
  - POST - shorten a url
    - Request-Body: { link: http://www.expandedurl.com }
    - Response-Body: { link: http://butly.com/key }
- / { key }
  - GET - get expanded url OR redirect?
    - Response-Body: { link: http://www.expandedurl.com }
- / trending
  - GET - get trending links
    - Response-Body: [ { link: http://www.expandedurl.com } , {...} , ... ]

### Database Schema

key|name     |type
---|---------|-----------
PK |id       |bigint(20)
   |orig_url |varchar(512)
   |short_url|varchar(45)
   |visits   |int(11)
   |created  |timestamp
