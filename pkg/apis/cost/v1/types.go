package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
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

// CostPolicy is the Schema for the costpolicies API
type CostPolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   CostPolicySpec   `json:"spec,omitempty"`
    Status CostPolicyStatus `json:"status,omitempty"`
}

// DeepCopyObject returns a generically typed copy of an object
func (in *CostPolicy) DeepCopyObject() runtime.Object {
    return in.DeepCopy()
}

// DeepCopy returns a deep copy of CostPolicy
func (in *CostPolicy) DeepCopy() *CostPolicy {
    if in == nil {
        return nil
    }
    out := new(CostPolicy)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *CostPolicy) DeepCopyInto(out *CostPolicy) {
    *out = *in
    out.TypeMeta = in.TypeMeta
    in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
    in.Spec.DeepCopyInto(&out.Spec)
    in.Status.DeepCopyInto(&out.Status)
}

// DeepCopyInto for CostPolicySpec
func (in *CostPolicySpec) DeepCopyInto(out *CostPolicySpec) {
    *out = *in
    in.OrphanedVolumes.DeepCopyInto(&out.OrphanedVolumes)
    in.IdleInstances.DeepCopyInto(&out.IdleInstances)
    in.TaggingPolicy.DeepCopyInto(&out.TaggingPolicy)
}

// DeepCopy for CostPolicySpec
func (in *CostPolicySpec) DeepCopy() *CostPolicySpec {
    if in == nil {
        return nil
    }
    out := new(CostPolicySpec)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyInto for OrphanedVolumePolicy
func (in *OrphanedVolumePolicy) DeepCopyInto(out *OrphanedVolumePolicy) {
    *out = *in
}

// DeepCopy for OrphanedVolumePolicy
func (in *OrphanedVolumePolicy) DeepCopy() *OrphanedVolumePolicy {
    if in == nil {
        return nil
    }
    out := new(OrphanedVolumePolicy)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyInto for IdleInstancePolicy
func (in *IdleInstancePolicy) DeepCopyInto(out *IdleInstancePolicy) {
    *out = *in
}

// DeepCopy for IdleInstancePolicy
func (in *IdleInstancePolicy) DeepCopy() *IdleInstancePolicy {
    if in == nil {
        return nil
    }
    out := new(IdleInstancePolicy)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyInto for TaggingPolicy
func (in *TaggingPolicy) DeepCopyInto(out *TaggingPolicy) {
    *out = *in
    if in.RequiredTags != nil {
        in, out := &in.RequiredTags, &out.RequiredTags
        *out = make([]string, len(*in))
        copy(*out, *in)
    }
}

// DeepCopy for TaggingPolicy
func (in *TaggingPolicy) DeepCopy() *TaggingPolicy {
    if in == nil {
        return nil
    }
    out := new(TaggingPolicy)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyInto for CostPolicyStatus
func (in *CostPolicyStatus) DeepCopyInto(out *CostPolicyStatus) {
    *out = *in
    in.LastScanTime.DeepCopyInto(&out.LastScanTime)
}

// DeepCopy for CostPolicyStatus
func (in *CostPolicyStatus) DeepCopy() *CostPolicyStatus {
    if in == nil {
        return nil
    }
    out := new(CostPolicyStatus)
    in.DeepCopyInto(out)
    return out
}

//+kubebuilder:object:root=true

// CostPolicyList contains a list of CostPolicy
type CostPolicyList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []CostPolicy `json:"items"`
}

// DeepCopyObject returns a generically typed copy of an object
func (in *CostPolicyList) DeepCopyObject() runtime.Object {
    return in.DeepCopy()
}

// DeepCopy returns a deep copy of CostPolicyList
func (in *CostPolicyList) DeepCopy() *CostPolicyList {
    if in == nil {
        return nil
    }
    out := new(CostPolicyList)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyInto copies all properties into another CostPolicyList
func (in *CostPolicyList) DeepCopyInto(out *CostPolicyList) {
    *out = *in
    out.TypeMeta = in.TypeMeta
    in.ListMeta.DeepCopyInto(&out.ListMeta)
    if in.Items != nil {
        in, out := &in.Items, &out.Items
        *out = make([]CostPolicy, len(*in))
        for i := range *in {
            (*in)[i].DeepCopyInto(&(*out)[i])
        }
    }
}
