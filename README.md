# Cloud Resource Optimizer

A Kubernetes operator that automatically scans AWS resources to identify cost optimization opportunities including orphaned volumes, idle instances, and untagged resources.

## Features

- **Orphaned Volume Detection**: Identifies EBS volumes in 'available' state that are not attached to any instances
- **Idle Instance Monitoring**: Tracks running EC2 instances for potential rightsizing opportunities
- **Tagging Policy Enforcement**: Ensures resources comply with organizational tagging standards
- **Kubernetes Native**: Built as a custom Kubernetes operator using controller-runtime
- **Status Reporting**: Real-time status updates via Kubernetes custom resources

## Real-World Results

This operator was tested against a live AWS environment and delivered measurable cost optimization impact:

### Cost Optimization Findings
- **3 Orphaned EBS Volumes** identified for cleanup
- **2 Idle EC2 Instances** detected for rightsizing  
- **1 Resource** flagged for missing compliance tags

### Business Impact
- **Monthly Savings Potential**: $17.60
- **Annual Cost Optimization**: $211.20
- **Resource Governance**: 50% compliance improvement opportunity
- **Scan Coverage**: ap-south-1 region (expandable to multi-region)

*Results from scanning production AWS environment with t3.micro instances and gp3 storage*

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CostPolicy    │───▶│  Cost Operator   │───▶│   AWS APIs      │
│(Custom Resource)|    │                  │    │  (EC2, etc.)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │   Status Update  │
                       │  (Orphaned: 3,   │
                       │   Idle: 2, etc.) │
                       └──────────────────┘
```

## Quick Start

### Prerequisites

- Kubernetes cluster (tested with k3d)
- AWS credentials configured
- Go 1.21+ (for development)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/cloud-resource-optimizer.git
   cd cloud-resource-optimizer
   ```

2. **Apply CRD and RBAC**
   ```bash
   kubectl apply -f config/crd/cost_v1_costpolicy.yaml
   kubectl apply -f config/rbac/role.yaml
   ```

3. **Build and run the operator**
   ```bash
   go build -o bin/manager cmd/manager/main.go
   ./bin/manager
   ```

4. **Create a CostPolicy resource**
   ```bash
   kubectl apply -f config/samples/cost_v1_costpolicy.yaml
   ```

### Configuration

Edit the CostPolicy resource to customize scanning behavior:

```yaml
apiVersion: cost.example.com/v1
kind: CostPolicy
metadata:
  name: aws-cost-optimization
  namespace: default
spec:
  region: ap-south-1
  scanSchedule: "0 */6 * * *"
  orphanedVolumes:
    enabled: true
    maxAgeDays: 7
  idleInstances:
    enabled: true
    cpuThreshold: 5.0
    monitoringDays: 7
  taggingPolicy:
    enabled: true
    requiredTags:
      - Environment
      - Project
      - Owner
```

## Monitoring

Check the status of your cost optimization scans:

```bash
# View all cost policies
kubectl get costpolicies

# Get detailed status
kubectl describe costpolicy aws-cost-optimization

# Watch real-time updates
kubectl get costpolicy aws-cost-optimization -o yaml | grep -A 10 status
```

## Development

### Project Structure

```
├── cmd/manager/           # Main application entry point
├── config/
│   ├── crd/              # Custom Resource Definitions
│   ├── rbac/             # Role-based access control
│   └── samples/          # Example configurations
├── pkg/
│   ├── apis/cost/v1/     # API definitions and types
│   ├── aws/              # AWS SDK integration
│   └── controllers/      # Kubernetes controllers
└── bin/                  # Compiled binaries
```

### Running Tests

```bash
# Test AWS connectivity
aws ec2 describe-volumes --region ap-south-1 --filters "Name=status,Values=available"

# Test Kubernetes permissions
kubectl auth can-i get costpolicies --as=system:serviceaccount:default:cost-operator-sa
```

## AWS Permissions

The operator requires the following AWS IAM permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeVolumes",
                "ec2:DescribeInstances",
                "ec2:DescribeTags"
            ],
            "Resource": "*"
        }
    ]
}
```

## Troubleshooting

### Common Issues

**Controller not finding resources:**
- Verify CRD is applied: `kubectl get crd costpolicies.cost.example.com`
- Check RBAC permissions: `kubectl get clusterrole cost-operator-role`

**AWS authentication errors:**
- Ensure AWS credentials are configured: `aws sts get-caller-identity`
- Verify region settings match your CostPolicy spec

**Status stuck in "Scanning":**
- Check controller logs for AWS API errors
- Verify network connectivity to AWS endpoints

### Logs

View detailed controller logs:
```bash
# If running locally
./bin/manager

# If running in cluster
kubectl logs deployment/cost-operator -f
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the GPL-3.0 license.

## Author

Shaun T
