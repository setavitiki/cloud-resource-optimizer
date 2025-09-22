package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CostPolicySpec defines the desired state of CostPolicy
type CostPolicySpec struct {
    // AWS region to scan
    Region string `json:"region,omitempty"`
    
    // Scan configuration
    ScanSchedule string `json:"scanSchedule,omitempty"` // Cron format
    
    // Cleanup policies
    OrphanedVolumes OrphanedVolumePolicy `json:"orphanedVolumes,omitempty"`
    IdleInstances   IdleInstancePolicy   `json:"idleInstances,omitempty"`
    TaggingPolicy   TaggingPolicy        `json:"taggingPolicy,omitempty"`
}

type OrphanedVolumePolicy struct {
    Enabled    bool `json:"enabled"`
    AutoDelete bool `json:"autoDelete,omitempty"`
    MaxAgeDays int  `json:"maxAgeDays,omitempty"`
}

type IdleInstancePolicy struct {
    Enabled         bool    `json:"enabled"`
    CPUThreshold    float64 `json:"cpuThreshold,omitempty"`
    MonitoringDays  int     `json:"monitoringDays,omitempty"`
    AutoStop        bool    `json:"autoStop,omitempty"`
}

type TaggingPolicy struct {
    Enabled      bool     `json:"enabled"`
    RequiredTags []string `json:"requiredTags,omitempty"`
}

// CostPolicyStatus defines the observed state of CostPolicy
type CostPolicyStatus struct {
    LastScanTime      metav1.Time `json:"lastScanTime,omitempty"`
    OrphanedVolumes   int         `json:"orphanedVolumes,omitempty"`
    IdleInstances     int         `json:"idleInstances,omitempty"`
    UntaggedResources int         `json:"untaggedResources,omitempty"`
    TotalSavings      string      `json:"totalSavings,omitempty"`
    Phase             string      `json:"phase,omitempty"`
    Message           string      `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.region`
//+kubebuilder:printcolumn:name="Orphaned Volumes",type=integer,JSONPath=`.status.orphanedVolumes`
//+kubebuilder:printcolumn:name="Idle Instances",type=integer,JSONPath=`.status.idleInstances`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// CostPolicy is the Schema for the costpolicies API
type CostPolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   CostPolicySpec   `json:"spec,omitempty"`
    Status CostPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CostPolicyList contains a list of CostPolicy
type CostPolicyList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []CostPolicy `json:"items"`
}
