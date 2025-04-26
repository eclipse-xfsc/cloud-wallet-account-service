# Account Service

The Account Service is a API gateway service designed to be an entry point for cloud PCM backend.

## API Documentation
The API documentation is written in Swagger [Swagger Web UI](docs/swagger.json)

It can be accessed in local docker-compose environment [swagger_ui](http://localhost:8080/swagger/index.html).

It is generated using [swagger generation tool](https://github.com/swaggo/gin-swagger)

In case of changes in api definitions, it should be updated. Please run:
```cmd
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency
```
## Available Endpoints

- GET /v1/metrics/health
- GET /swagger/*any
- GET /v1/tenants/:tenantId/api/accounts/devices/list
- GET /v1/tenants/:tenantId/api/accounts/devices/link
- DELETE /v1/tenants/:tenantId/api/accounts/devices/:id
- POST /v1/tenants/:tenantId/api/accounts/devices/block/:id
- GET /v1/tenants/:tenantId/api/accounts/history/list
- GET /v1/tenants/:tenantId/api/accounts/kms/keyTypes
- GET /v1/tenants/:tenantId/api/accounts/kms/did/list
- POST /v1/tenants/:tenantId/api/accounts/kms/did/create
- DELETE /v1/tenants/:tenantId/api/accounts/kms/did/:kid
- GET /v1/tenants/:tenantId/api/accounts/settings/ui
- POST /v1/tenants/:tenantId/api/accounts/settings/ui
- GET /v1/tenants/:tenantId/api/accounts/credentials/list
- POST /v1/tenants/:tenantId/api/accounts/credentials/list
- GET /v1/tenants/:tenantId/api/accounts/credentials/history
- DELETE /v1/tenants/:tenantId/api/accounts/credentials/:id
- GET /v1/tenants/:tenantId/api/accounts/credentials/:id/revoke
- GET /v1/tenants/:tenantId/api/accounts/credentials/schemas
- POST /v1/tenants/:tenantId/api/accounts/credentials/issue
- PUT /v1/tenants/:tenantId/api/accounts/credentials/backup/:id/:bid
- GET /v1/tenants/:tenantId/api/accounts/credentials/backup/:id/:bid
- GET /v1/tenants/:tenantId/api/accounts/credentials/backup/link/:mode
- GET /v1/tenants/:tenantId/api/accounts/credentials/backup/all
- GET /v1/tenants/:tenantId/api/accounts/credentials/backup/latest
- DELETE /v1/tenants/:tenantId/api/accounts/credentials/backup/invalid
- DELETE /v1/tenants/:tenantId/api/accounts/credentials/backup/:bid
- PUT /v1/tenants/:tenantId/api/accounts/credentials/offers/create
- GET /v1/tenants/:tenantId/api/accounts/credentials/offers/list
- POST /v1/tenants/:tenantId/api/accounts/credentials/offers/:id/accept

## Dependency services

[docker-compose](./deployment/docker/docker-compose.yml)

Infrastructure
- [Postgres](https://www.postgresql.org/)
- [Nats](https://nats.io/)
- [Vault](https://www.vaultproject.io/)

Microservices
- [Storage service](https://gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/storage-service)
- [Credential verification service](https://gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/credential-verification-service)
- [Credential retrieval service](https://gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/credential-retrieval-service)
- [Issuer service](https://gitlab.eclipse.org/eclipse/xfsc/organisational-credential-manager-w-stack/credential-issuance)
- [Signer service](https://gitlab.eclipse.org/eclipse/xfsc/tsa/signer/-/tree/ocm-wstack?ref_type=heads)

## Running the Service

### In [Docker container](https://docs.docker.com/engine/install/) (recommended)

1. Setup
.env file or config.yaml can be used to define environment variables.

2. Build and run service

If necessary, copy to terminal corresponding command from [Makefile](makefile)

```make docker-compose-run```

### Locally

1. Setup
   Define environment variables in terminal
```
ACCOUNT_LISTENPORT=8080
ACCOUNT_SERVERMODE=production
ACCOUNT_NATS_WITHNATS=true
ACCOUNT_NATS_TIMEOUTINSEC=10s
ACCOUNT_NATS_URL=nats.nats.svc.cluster.local:4222
ACCOUNT_NATS_QUEUEGROUP=account-service
ACCOUNT_CLOUDEVENTS_TOPICS=didcomm-connector-invitation
ACCOUNT_MESSAGING_PROTOCOL=nats
ACCOUNT_MESSAGING_NATS_WITH_NATS=true
ACCOUNT_MESSAGING_NATS_URL=nats.nats.svc.cluster.local:4222
ACCOUNT_MESSAGING_NATS_TIMEOUTINSEC=10s
ACCOUNT_PROTOCOL=nats
ACCOUNT_DB_WITHDB=true
ACCOUNT_DB_DBTYPE=postgres
ACCOUNT_DB_HOST=testing-db-rw.postgres.svc.cluster.local
ACCOUNT_DB_PORT=5432
ACCOUNT_DB_USER=app
ACCOUNT_DB_PASSWORD=change me
ACCOUNT_DB_DBNAME=app
ACCOUNT_KEYCLOAK_URL=http://keycloak.keycloak.svc.cluster.local
ACCOUNT_KEYCLOAK_REALMNAME=react-keycloak
ACCOUNT_KEYCLOAK_TOKENTTL=250ns
ACCOUNT_KEYCLOAK_EXCLUDEENDPOINTS=/v1/tenants/:tenantId/api/accounts/credentials/backup/:id/:bid
ACCOUNT_STORAGE_URL=http://storage-service.default.svc.cluster.local:8080/v1/tenants/tenant_space/storage
ACCOUNT_STORAGE_WITHAUTH=false
ACCOUNT_BACKUPLINKTTL=300s
ACCOUNT_DIDCOMM_URL=http://didcomm-connector.default.svc.cluster.local:9090
ACCOUNT_SIGNER_URL=http://signer.default.svc.cluster.local:8080
ACCOUNT_CREDENTIALRETRIEVAL_URL=http://credential-retrieval-service.default.svc.cluster.local:8080/v1/tenants/tenant_space
ACCOUNT_CREDENTIALRETRIEVAL_OFFERTOPIC=offering
ACCOUNT_CREDENTIALVERIFIER_URL=http://credential-verification-service.default.svc.cluster.local:8080/v1/tenants/tenant_space/internal
VAULT_ADRESS=http://vault.vault.svc.cluster.local:8200
VAULT_TOKEN=mytoken
```

2. Build plugin

```cmd
mkdir etc
cd etc
git clone https://gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/plugins/hashicorp-vault-provider.git .
go mod download
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -mod=mod -buildmode=plugin -o /etc/plugins
```
3. Build and run service

```cmd
go mod download
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /microservice
./microservice
```

## Deploying the service

Deployment is managed using [Helm](https://helm.sh)

Charts are defined in [./deployment/helm](deployment/helm)

From root

```cmd
cd deployment
helm install account-service ./helm -n <your namespace | default> --kubeconfig <path to kubeconfig of necessary cluster>
```