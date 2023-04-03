# Eliona App for Glutz Devices
  
This [Eliona app for Glutz Devices](https://github.com/eliona-smart-building-assistant/glutz-app) allows the opening and closing of doors fitted with Glutz devices by enabling the data transfer between these devices and the Eliona environment. 

## Configuration

The app needs environment variables and database tables for configuration. To edit the database tables the app provides its own API access.


### Registration in Eliona ###

To start and initialize an app in an Eliona environment, the app has to registered in Eliona. For this, an entry in the database table `public.eliona_app` is necessary.


### Environment variables

- `APPNAME`: must be set to `glutz`. Some resources use this name to identify the app inside an Eliona environment.

- `CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db). Otherwise, the app can't be initialized and started. (e.g. `postgres://user:pass@localhost:5432/iot`)

- `API_ENDPOINT`:  configures the endpoint to access the [Eliona API v2](https://github.com/eliona-smart-building-assistant/eliona-api). Otherwise, the app can't be initialized and started. (e.g. `http://api-v2:3000/v2`)

- `API_TOKEN`: defines the secret to authenticate the app and access the API. 

- `API_SERVER_PORT`(optional): define the port the API server listens. The default value is Port `3000`. <mark>Todo: Decide if the app needs its own API. If so, an API server have to implemented and the port have to be configurable.</mark>

- `LOG_LEVEL`(optional): defines the minimum level that should be [logged](https://github.com/eliona-smart-building-assistant/go-utils/blob/main/log/README.md). Not defined the default level is `info`.

### Database tables ###

The app requires configuration data that remains in the database. In order to store the data, the app creates its own database schema `glutz` during initialization. To modify and handle the configuration data the app provides an API access. Take a look at the [API specification](https://github.com/eliona-smart-building-assistant/glutz-app/blob/develop/openapi.yaml) to see how the configuration tables should be used.

- `glutz.config`: contains the Glutz API endpoints. Each row contains the specification of one endpoint (i.e config id, username, password, polling interval etc.)

- `glutz.spaces`: contains the mapping from each device (uniquely defined by its configuration-, project- and device- id) to an eliona asset. Each row contains the specification of one endpoint(i.e config id, username, password, polling interval etc.) The app collects and writes data separately for each configured project. The mapping is created automatically by the app.

**Generation**: to generate access method to database see Generation section below.


## References

### App API ###

The Glutz app provides its own API to access configuration data and other functions. The full description of the API is defined in the `openapi.yaml` OpenAPI definition file.

- [API Reference](https://github.com/eliona-smart-building-assistant/glutz-app/blob/develop/openapi.yaml) shows details of the API

**Generation**: to generate api server stub see Generation section below.


### Eliona Assets ###

The app creates necessary asset types and attributes during initialization. See [eliona/asset-type-glutz_device.json](eliona/asset-type-glutz_device.json) for details.

Each Glutz device is automatically mapped to an asset with atrributes of the subtype `Input`, `Info` and `Output`. The Glutz app writes input (e.g battery level, number of openings) and info (e.g building, room, open) data for each Glutz device to the eliona database and reads output data (openable) from Eliona.


## Tools

### Generate API server stub ###

For the API server the [OpenAPI Generator](https://openapi-generator.tech/docs/generators/openapi-yaml) for go-server is used to generate a server stub. The easiest way to generate the server files is to use one of the predefined generation scripts, which use the OpenAPI Generator Docker image.

```
.\generate-api-server.cmd # Windows
./generate-api-server.sh # Linux
```

### Generate Database access ###

For the database access [SQLBoiler](https://github.com/volatiletech/sqlboiler) is used. The easiest way to generate the database files is to use one of the predefined generation scripts, which use the SQLBoiler implementation. Please note that the database connection in the `sqlboiler.toml` file have to be configured.

```
.\generate-db.cmd # Windows
./generate-db.sh # Linux
```
