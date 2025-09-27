package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Scanner struct {
	ec2Client *ec2.Client
	region    string
}

func NewScanner(region string) (*Scanner, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &Scanner{
		ec2Client: ec2.NewFromConfig(cfg),
		region:    region,
	}, nil
}

func (s *Scanner) ScanOrphanedVolumes(ctx context.Context) ([]types.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("status"), // Changed from "state" to "status"
				Values: []string{"available"},
			},
		},
	}

	result, err := s.ec2Client.DescribeVolumes(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Volumes, nil
}

func (s *Scanner) ScanIdleInstances(ctx context.Context) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	}

	result, err := s.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	var instances []types.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances, nil
}

func (s *Scanner) ScanUntaggedResources(ctx context.Context, requiredTags []string) (int, error) {
	instances, err := s.ScanIdleInstances(ctx)
	if err != nil {
		return 0, err
	}

	untagged := 0
	for _, instance := range instances {
		tagMap := make(map[string]string)
		for _, tag := range instance.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		for _, requiredTag := range requiredTags {
			if _, exists := tagMap[requiredTag]; !exists {
				untagged++
				break
			}
		}
	}

	return untagged, nil
}