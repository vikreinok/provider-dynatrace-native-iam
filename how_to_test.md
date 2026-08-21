# How to Test the Native Dynatrace Crossplane Provider (`provider-dynatrace-native`)

This guide explains how to run manual and automated integration tests for the native Dynatrace Crossplane provider using your Dynatrace account or trial credentials.

Testing is done by running the provider locally or in-cluster while utilizing a local Kubernetes cluster (`kind`) as the API control plane.

---

## 1. Prerequisites

1. **Docker**: Ensure Docker daemon / Docker Desktop is running.
2. **Kind**: Kubernetes in Docker (`brew install kind`).
3. **Go**: Go version `>=1.24` (`brew install go`).
4. **Kubectl**: Kubernetes CLI tool (`brew install kubectl`).
5. **Helm**: Kubernetes package manager (`brew install helm`).

---

## 2. Credentials Configuration (`secret.yaml`)

The provider consumes credentials from a standard Kubernetes Secret referenced by `ClusterProviderConfig` (or namespaced `ProviderConfig`). 

The JSON credentials format is 100% backward compatible with the older Upjet provider (`provider-dynatrace-iam`), supporting both `iam_*` and `dt_*` key conventions.

### Create `secret.yaml` (Ignored by Git)

Create a `secret.yaml` file in the root of the repository. This file is excluded by `.gitignore` to prevent credentials from being committed:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dynatrace-creds
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "dt_env_url": "https://<your-environment-id>.live.dynatrace.com",
      "iam_account_id": "<your-iam-account-id>",
      "iam_client_id": "<your-iam-client-id>",
      "iam_client_secret": "<your-iam-client-secret>"
    }
```

Apply the secret and provider configuration:
```bash
kubectl apply -f secret.yaml
kubectl apply -f examples/provider/config.yaml
```

---

## 3. Running Automated Live API Integration Tests

To test pure REST client interactions against the live Dynatrace OAuth2 and IAM endpoints directly without Kubernetes:

```bash
# Option A: Export environment variables
export DT_ACCOUNT_ID="<your-iam-account-id>"
export DT_CLIENT_ID="<your-iam-client-id>"
export DT_CLIENT_SECRET="<your-iam-client-secret>"
export DT_ENV_URL="https://<your-environment-id>.live.dynatrace.com"

go test -tags=integration -v ./test/integration/...

# Option B: Or simply place your credentials in secret.yaml; the tests will automatically discover them!
go test -tags=integration -v ./test/integration/...
```

---

## 4. Step-by-Step Manual Testing Workflow

### Step 4.1: Create Local Kind Cluster & Install Crossplane

```bash
# Create kind cluster
kind create cluster --name dynatrace-iam-test

# Install Crossplane
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm upgrade --install crossplane crossplane-stable/crossplane \
    --namespace crossplane-system \
    --create-namespace \
    --wait
```

### Step 4.2: Apply Native CRDs

Apply the native Dynatrace CRDs from `package/crds/`:

```bash
kubectl apply -f package/crds/
```

Verify the registered CRDs:
```bash
kubectl get crds | grep dynatrace
```

### Step 4.3: Apply Secret & ClusterProviderConfig

```bash
# Apply your gitignored credentials secret
kubectl apply -f secret.yaml

# Apply provider config
kubectl apply -f examples/provider/config.yaml
```

### Step 4.4: Run the Provider Controller Locally

You can run the provider controller out-of-cluster locally using either of the following approaches:

#### Option A: Using Makefile (Standard Project Workflow)

```bash
export PATH=$PATH:$HOME/go/bin
unset GOROOT
export GOTOOLCHAIN=local
make run
```
*Note: `make run` compiles the provider binary and executes it with `--debug` logging.*

#### Option B: Direct Go Execution

```bash
# Direct run
go run cmd/provider/main.go --debug

# Or build and run binary
go build -o /tmp/provider ./cmd/provider
/tmp/provider --debug
```

---

### Step 4.5: Create Managed IAM Resources

Apply sample resources from `examples/iam/`:

#### Example 1: Dynatrace IAM Group
```yaml
# examples/iam/group.yaml
apiVersion: iam.dynatrace.crossplane.io/v1alpha1
kind: Group
metadata:
  name: platform-engineers
spec:
  forProvider:
    name: "Platform Engineers"
    description: "Platform engineering team managed by Crossplane"
    federatedAttributeValues:
      - "platform-engineers-saml"
  providerConfigRef:
    name: default
```

#### Example 2: Dynatrace IAM Policy
```yaml
# examples/iam/policy.yaml
apiVersion: iam.dynatrace.crossplane.io/v1alpha1
kind: Policy
metadata:
  name: settings-reader-policy
spec:
  forProvider:
    name: "Settings Reader Policy"
    description: "Allows reading settings schemas and objects"
    statementQuery: 'ALLOW settings:objects:read, settings:schemas:read;'
    account: "<your-iam-account-id>"
  providerConfigRef:
    name: default
```

#### Example 3: Dynatrace IAM CostCenter
```yaml
# examples/iam/costcenter.yaml
apiVersion: iam.dynatrace.crossplane.io/v1alpha1
kind: CostCenter
metadata:
  name: platform-eng-cc
spec:
  forProvider:
    costCenter: "Platform Engineering"
  providerConfigRef:
    name: default
```

Apply the manifests:
```bash
kubectl apply -f examples/iam/group.yaml
kubectl apply -f examples/iam/policy.yaml
kubectl apply -f examples/iam/costcenter.yaml
```

---

### Step 4.6: Observe Reconciliation & Status

Check the status of your resources:

```bash
kubectl get group,policy,costcenter,policybindingsv2,user
```

Example Output:
```text
NAME                                                  READY   SYNCED   EXTERNAL-NAME                          AGE
group.iam.dynatrace.crossplane.io/platform-engineers   True    True     b35f4a41-5a4d-45ea-b7a6-514ab5660fb7   15s

NAME                                                        READY   SYNCED   EXTERNAL-NAME                          AGE
policy.iam.dynatrace.crossplane.io/settings-reader-policy   True    True     89e4c1a2-3f12-4211-9a10-123456789abc   15s

NAME                                                     READY   SYNCED   EXTERNAL-NAME          AGE
costcenter.iam.dynatrace.crossplane.io/platform-eng-cc   True    True     Platform Engineering   15s
```

Inspect details and observable status fields:
```bash
kubectl describe group platform-engineers
```

---

### Step 4.7: Verify Drift Correction

1. Log into your Dynatrace Account / Environment web console.
2. Manually modify the description of the created group or policy.
3. Within the next reconciliation cycle (1 minute), verify the provider detects the drift and updates Dynatrace back to the desired specification declared in Kubernetes.

---

### Step 4.8: Clean Teardown

Delete the managed resources (which instructs the provider to cleanly remove the external entities in Dynatrace):

```bash
kubectl delete -f examples/iam/group.yaml
kubectl delete -f examples/iam/policy.yaml
kubectl delete -f examples/iam/costcenter.yaml
```

Stop the local provider process (`Ctrl+C`), and optionally delete the test cluster:

```bash
kind delete cluster --name dynatrace-iam-test
```
