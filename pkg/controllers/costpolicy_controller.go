package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
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
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main Kubernetes reconciliation loop
func (r *CostPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Starting reconciliation", "namespacedName", req.NamespacedName)

	var costPolicy costv1.CostPolicy
	if err := r.Get(ctx, req.NamespacedName, &costPolicy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("CostPolicy resource not found, may have been deleted", "namespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CostPolicy", "namespacedName", req.NamespacedName)
		return ctrl.Result{}, err
	}

	log.Info("Successfully retrieved CostPolicy", "name", costPolicy.Name, "region", costPolicy.Spec.Region)

	// Initialize AWS scanner
	log.Info("Initializing AWS scanner", "region", costPolicy.Spec.Region)
	scanner, err := aws.NewScanner(costPolicy.Spec.Region)
	if err != nil {
		log.Error(err, "Failed to create AWS scanner", "region", costPolicy.Spec.Region)
		
		// Update status with error
		costPolicy.Status.Phase = "Failed"
		costPolicy.Status.Message = fmt.Sprintf("Failed to create AWS scanner: %v", err)
		if statusErr := r.Status().Update(ctx, &costPolicy); statusErr != nil {
			log.Error(statusErr, "Failed to update status after scanner creation failure")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
	}

	log.Info("AWS scanner created successfully")

	// Update status to scanning
	costPolicy.Status.Phase = "Scanning"
	costPolicy.Status.Message = "Scanning AWS resources..."
	costPolicy.Status.LastScanTime = metav1.Time{Time: time.Now()}
	
	log.Info("Updating status to Scanning phase")
	if err := r.Status().Update(ctx, &costPolicy); err != nil {
		log.Error(err, "Failed to update status to Scanning phase")
		return ctrl.Result{}, err
	}

	// Perform scans
	var orphanedCount, idleCount, untaggedCount int

	if costPolicy.Spec.OrphanedVolumes.Enabled {
		log.Info("Scanning for orphaned volumes")
		volumes, err := scanner.ScanOrphanedVolumes(ctx)
		if err != nil {
			log.Error(err, "Failed to scan orphaned volumes")
		} else {
			orphanedCount = len(volumes)
			log.Info("Orphaned volumes scan completed", "count", orphanedCount)
		}
	}

	if costPolicy.Spec.IdleInstances.Enabled {
		log.Info("Scanning for idle instances")
		instances, err := scanner.ScanIdleInstances(ctx)
		if err != nil {
			log.Error(err, "Failed to scan idle instances")
		} else {
			idleCount = len(instances)
			log.Info("Idle instances scan completed", "count", idleCount)
		}
	}

	if costPolicy.Spec.TaggingPolicy.Enabled {
		log.Info("Scanning for untagged resources", "requiredTags", costPolicy.Spec.TaggingPolicy.RequiredTags)
		count, err := scanner.ScanUntaggedResources(ctx, costPolicy.Spec.TaggingPolicy.RequiredTags)
		if err != nil {
			log.Error(err, "Failed to scan untagged resources")
		} else {
			untaggedCount = count
			log.Info("Untagged resources scan completed", "count", untaggedCount)
		}
	}

	// Update final status
	log.Info("Updating final status", "orphaned", orphanedCount, "idle", idleCount, "untagged", untaggedCount)
	
	costPolicy.Status.OrphanedVolumes = orphanedCount
	costPolicy.Status.IdleInstances = idleCount
	costPolicy.Status.UntaggedResources = untaggedCount
	costPolicy.Status.Phase = "Ready"
	costPolicy.Status.Message = "Scan completed successfully"
	
	if err := r.Status().Update(ctx, &costPolicy); err != nil {
		log.Error(err, "Failed to update final status")
		return ctrl.Result{}, err
	}

	log.Info("Reconciliation completed successfully", "orphaned", orphanedCount, "idle", idleCount, "untagged", untaggedCount)

	// Requeue based on schedule (default 1 hour)
	return ctrl.Result{RequeueAfter: time.Hour}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *CostPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&costv1.CostPolicy{}).
		Complete(r)
}