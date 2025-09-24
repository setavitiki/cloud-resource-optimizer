package controllers

import (
    "context"
    "time"

    "k8s.io/apimachinery/pkg/runtime"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    costv1 "github.com/setavitiki/cloud-resource-optimizer/pkg/apis/cost/v1"
    "github.com/setavitiki/cloud-resource-optimizer/pkg/aws"
)

// CostPolicyReconciler reconciles a CostPolicy object
type CostPolicyReconciler struct {
    client.Client           // cached client (used for watches, status updates)
    APIReader client.Reader // direct client (used for guaranteed fresh GETs)
    Scheme    *runtime.Scheme
}

// Reconcile is part of the main Kubernetes reconciliation loop
func (r *CostPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    var costPolicy costv1.CostPolicy
    // Use APIReader instead of cached client for fetching
    if err := r.APIReader.Get(ctx, req.NamespacedName, &costPolicy); err != nil {
        log.Error(err, "failed to get CostPolicy")
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    log.Info("Reconciling CostPolicy", "name", costPolicy.Name, "region", costPolicy.Spec.Region)

    // Initialize AWS scanner
    scanner, err := aws.NewScanner(costPolicy.Spec.Region)
    if err != nil {
        log.Error(err, "failed to create AWS scanner")
        costPolicy.Status.Phase = "Failed"
        costPolicy.Status.Message = err.Error()
        _ = r.Status().Update(ctx, &costPolicy)
        return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
    }

    // Update status to scanning
    costPolicy.Status.Phase = "Scanning"
    costPolicy.Status.Message = "Scanning AWS resources..."
    costPolicy.Status.LastScanTime = metav1.Time{Time: time.Now()}

    if err := r.Status().Update(ctx, &costPolicy); err != nil {
        return ctrl.Result{}, err
    }

    // Perform scans
    var orphanedCount, idleCount, untaggedCount int

    if costPolicy.Spec.OrphanedVolumes.Enabled {
        volumes, err := scanner.ScanOrphanedVolumes(ctx)
        if err != nil {
            log.Error(err, "failed to scan orphaned volumes")
        } else {
            orphanedCount = len(volumes)
        }
    }

    if costPolicy.Spec.IdleInstances.Enabled {
        instances, err := scanner.ScanIdleInstances(ctx)
        if err != nil {
            log.Error(err, "failed to scan idle instances")
        } else {
            idleCount = len(instances)
        }
    }

    if costPolicy.Spec.TaggingPolicy.Enabled {
        count, err := scanner.ScanUntaggedResources(ctx, costPolicy.Spec.TaggingPolicy.RequiredTags)
        if err != nil {
            log.Error(err, "failed to scan untagged resources")
        } else {
            untaggedCount = count
        }
    }

    // Update final status
    costPolicy.Status.OrphanedVolumes = orphanedCount
    costPolicy.Status.IdleInstances = idleCount
    costPolicy.Status.UntaggedResources = untaggedCount
    costPolicy.Status.Phase = "Ready"
    costPolicy.Status.Message = "Scan completed successfully"

    if err := r.Status().Update(ctx, &costPolicy); err != nil {
        return ctrl.Result{}, err
    }

    log.Info("Scan completed", "orphaned", orphanedCount, "idle", idleCount, "untagged", untaggedCount)

    // Requeue based on schedule (default 1 hour)
    return ctrl.Result{RequeueAfter: time.Hour}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *CostPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&costv1.CostPolicy{}).
        Complete(r)
}