# Basic Requirements

## Basic Functions
- Accept metrics from upmaster-agent and store them into database
- Dispatch HTTP Endpoints to upmaster-agent
- Provide API for upmaster-frontend to visualize data
- Alert user when certain contraint meets

## Platform Requirements

Require Go >= 1.15.8

## Commit Message Convention

Use AngularJS style commit message.

# Development

## Rough Structure

## Detailed Design
### Database Table Design

Two database is used in this project: InfluxDB and SQLite. Since UpMaster is designed for small teams, account information should be handled well by SQLite, while time serires data is stored in InfluxDB.

#### SQLite Table Design

**Users**
| ID       | Username | Alias  | Password      | Email  | IsAdmin | Endpoints   | Alerts      | AlertChannels |
| -------- | -------- | ------ | ------------- | ------ | ------- | ----------- | ----------- | ------------- |
| Main Key | String   | String | Hashed String | String | Bool    | One-to-many | One-to-many | One-to-many   |

**Endpoints**
| ID       | Name   | UserID      | URL    | Interval     | IsEnabled | IsVisible | Alerts      |
| -------- | ------ | ----------- | ------ | ------------ | --------- | --------- | ----------- |
| Main Key | String | Foreign Key | String | Int (Second) | Bool      | Bool      | One-to-many |

**AlertChannels**
| ID       | Name   | UserID      | Type             | Config | Alerts      | IsEnabled |
| -------- | ------ | ----------- | ---------------- | ------ | ----------- | --------- |
| Main Key | String | Foreign Key | Int (With Marco) | byte[] | One-to-many | Bool      |

**Alerts**
| ID       | UserID      | AlertChannelID | Status                  |
| -------- | ----------- | -------------- | ----------------------- |
| Main Key | Foreign Key | Foreign Key    | Int (Alerting/Resolved) |

**Configs**
Used to store dynamic InfluxDB configuration
| ID       | Key    | Value                  |
| -------- | ------ | ---------------------- |
| Main Key | String | Byte[] (After Marshal) |

**OAuth**

Used to provide storage for OAuth Server

#### InfluxDB Design

Measurement: **upstatus**

| Time | IsUp             | Node             | EndpointID    |
| ---- | ---------------- | ---------------- | ------------- |
| -    | Bool (Field Key) | String (Tag Key) | Int (Tag Key) |

### Initialization Process

The initialization process is designed to be **idempotent**. It collects configuration from config file or environment variable and reconfigure UpMaster.

The process will do the following steps:
- Initialize SQLite: Using GORM operations to do database initialization, idempotent is achieved by which.
- Update Admin Info: Create/Update a admin user according to config file.
- Initialize InfluxDB: **influxdb-client-go** by InfluxDB Official is used as client. Database connection info is retrieved from SQLite. The database should already be create before this step. Measurement `up_status` will be created if not exists.

### Alert Module Design

If any alert channel is created, a `StatusChecker` is created as a goroutine. It periodically poll data from InfluxDB and calculate if an endpoint is down. Then `StatusChecker` send an alert to all configured alert channel.

The status change is decided by a 'sliding window' algorithm, which consider endpoint down if all the points in the window is down. The same strategy is applied to nodes, that is to say, a endpoint is considered down only when all the agents report endpoint down.

### Authencation Module Design

OAuth Server

Backend has a OAuth Server library, which uses oauth table in database. The OAuth library is configured to use `JWT` to generate `accessToken`. This `accessToken` must have a field identifying different client types. Token for frontend is set to expired in a short time like 30mins, token for upmaster-agent is set to expired in a longer time, such as 12h. 

Casbin RBAC

Token for frontend cannot write time series data, while token for agent cannot read any data.

Frontend

`accessToken` and `refreshToken` is used in frontend. Both of them will be save in Session Cookie instead of Permanent Cookie if `keep me login` option is not checked. Frontend will verify if token is still valid by doing following things. If one of this failed, frontend will mark user as logout and redirect user to login page. 

1. Check `exp` field in `accessToken` and `refreshToken`.
2. Send a request to `/users/<username>` to get user object and verify if token is valid at the same time.

Before every request, frontend will check expire time of `accessToken` and request `/auth/refresh` API with refreshToken first based on the situation.

Agent

OAuth 2.0 is intended to be used. Agent reads `client_id` and `client_secret` from config file or environment variable, and use them to get a `token` from backend. Then the `token` is used in api authentication. The `token` my expired, but it will be automatically renew by OAuth library. As a result, `.Token()` method must be called every time a new api call is conducted.

### API Design Principle
API should be prefixed with version number, such as `v1`.

API can be prefixed with certain prefix to avoid conflict with frontend. For example `api/v1`.

The following design should be prefixed with `api/v1`:

**Authentication API** /auth/

POST `/auth/login` return JWT token for frontend

POST `/auth/reset` Used to send verification token
PUT `/auth/reset` Used by reset password (including update password)

OAuth Server at `/oauth`

**Endpoint API** /endpoints required Admin (Except GET /)

GET `/endpoints` Get all endpoints info (Used by agent)

Other endpoint direct operation

**Status API** /endpoints/<endpoint_id>/status

PUT `/endpoints/<endpoint_id>/status` is used by agent to write time series data

GET `/endpoints/<endpoint_id>/status` is used by frontend (Certain endpoint page)

**User API** /users

GET `/users` Admin

User Permission is restricted within `/users/<username>` and `/users/<username>`

PUT DELETE `/users/<username>` Used by frontend admin page

GET `/users/<username>/endpoints` Used by frontend user

POST `/users/<username>/endpoints` Used by frontend user

PUT DELETE `/users/<username>/endpoints/<endpoint_id>` Used by frontend user

GET `/users/<username>/status` Used by frontend public page

**AlertChannel API** /alertchannels required Admin

GET `/alertchannels` Get all alert channels.

PUT DELETE `/alertchannels` Create/Delete a alert channel.

PUT `/alertchannels/<id>` Update a alert channel.

**Status API** /status

PUT `/status` is used by agent to write time series data

GET `/status/<endpoint_id>` is used by frontend (Certain endpoint page)


## Test Methods
