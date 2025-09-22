import boto3
import json
from datetime import datetime, timedelta
from prometheus_client import Counter, Gauge, Histogram
import logging

# Prometheus metrics
orphaned_volumes_gauge = Gauge('aws_orphaned_volumes_total', 'Number of orphaned EBS volumes')
idle_instances_gauge = Gauge('aws_idle_instances_total', 'Number of idle EC2 instances')
untagged_resources_gauge = Gauge('aws_untagged_resources_total', 'Number of untagged resources')
scan_duration_histogram = Histogram('aws_scan_duration_seconds', 'AWS scan duration', ['resource_type'])

class AWSResourceScanner:
    def __init__(self, region='us-east-1'):
        self.region = region
        self.ec2 = boto3.client('ec2', region_name=region)
        self.cloudwatch = boto3.client('cloudwatch', region_name=region)
        
    def scan_orphaned_volumes(self):
        """Find EBS volumes not attached to any instance"""
        with scan_duration_histogram.labels(resource_type='ebs').time():
            volumes = self.ec2.describe_volumes(
                Filters=[{'Name': 'state', 'Values': ['available']}]
            )
            
            orphaned_volumes = []
            for volume in volumes['Volumes']:
                orphaned_volumes.append({
                    'volume_id': volume['VolumeId'],
                    'size': volume['Size'],
                    'created_date': volume['CreateTime'].isoformat(),
                    'volume_type': volume['VolumeType']
                })
            
            orphaned_volumes_gauge.set(len(orphaned_volumes))
            return orphaned_volumes
    
    def scan_idle_instances(self, cpu_threshold=5.0, days_back=7):
        """Find EC2 instances with low CPU utilization"""
        with scan_duration_histogram.labels(resource_type='ec2').time():
            instances = self.ec2.describe_instances(
                Filters=[{'Name': 'instance-state-name', 'Values': ['running']}]
            )
            
            idle_instances = []
            end_time = datetime.utcnow()
            start_time = end_time - timedelta(days=days_back)
            
            for reservation in instances['Reservations']:
                for instance in reservation['Instances']:
                    instance_id = instance['InstanceId']
                    
                    # Get CPU utilization metrics
                    try:
                        cpu_metrics = self.cloudwatch.get_metric_statistics(
                            Namespace='AWS/EC2',
                            MetricName='CPUUtilization',
                            Dimensions=[{'Name': 'InstanceId', 'Value': instance_id}],
                            StartTime=start_time,
                            EndTime=end_time,
                            Period=3600,  # 1 hour
                            Statistics=['Average']
                        )
                        
                        if cpu_metrics['Datapoints']:
                            avg_cpu = sum(dp['Average'] for dp in cpu_metrics['Datapoints']) / len(cpu_metrics['Datapoints'])
                            
                            if avg_cpu < cpu_threshold:
                                idle_instances.append({
                                    'instance_id': instance_id,
                                    'instance_type': instance['InstanceType'],
                                    'avg_cpu_utilization': round(avg_cpu, 2),
                                    'launch_time': instance['LaunchTime'].isoformat()
                                })
                    except Exception as e:
                        logging.warning(f"Could not get CPU metrics for {instance_id}: {str(e)}")
            
            idle_instances_gauge.set(len(idle_instances))
            return idle_instances
    
    def scan_untagged_resources(self):
        """Find resources without proper tags"""
        with scan_duration_histogram.labels(resource_type='tags').time():
            untagged_resources = []
            
            # Check EC2 instances
            instances = self.ec2.describe_instances()
            for reservation in instances['Reservations']:
                for instance in reservation['Instances']:
                    tags = instance.get('Tags', [])
                    tag_keys = [tag['Key'] for tag in tags]
                    
                    required_tags = ['Environment', 'Project', 'Owner']
                    missing_tags = [tag for tag in required_tags if tag not in tag_keys]
                    
                    if missing_tags:
                        untagged_resources.append({
                            'resource_type': 'EC2',
                            'resource_id': instance['InstanceId'],
                            'missing_tags': missing_tags
                        })
            
            # Check EBS volumes
            volumes = self.ec2.describe_volumes()
            for volume in volumes['Volumes']:
                tags = volume.get('Tags', [])
                tag_keys = [tag['Key'] for tag in tags]
                
                required_tags = ['Environment', 'Project']
                missing_tags = [tag for tag in required_tags if tag not in tag_keys]
                
                if missing_tags:
                    untagged_resources.append({
                        'resource_type': 'EBS',
                        'resource_id': volume['VolumeId'],
                        'missing_tags': missing_tags
                    })
            
            untagged_resources_gauge.set(len(untagged_resources))
            return untagged_resources
    
    def full_scan(self):
        """Perform complete resource scan"""
        results = {
            'timestamp': datetime.utcnow().isoformat(),
            'region': self.region,
            'orphaned_volumes': self.scan_orphaned_volumes(),
            'idle_instances': self.scan_idle_instances(),
            'untagged_resources': self.scan_untagged_resources()
        }
        
        logging.info(f"Scan completed: {len(results['orphaned_volumes'])} orphaned volumes, "
                    f"{len(results['idle_instances'])} idle instances, "
                    f"{len(results['untagged_resources'])} untagged resources")
        
        return results
