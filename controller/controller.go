/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"time"
	"strconv"
	"strings"
	"regexp"

	"golang.org/x/time/rate"


	//appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	//appsinformers "k8s.io/client-go/informers/apps/v1"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	//appslisters "k8s.io/client-go/listers/apps/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	samplev1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"
	samplescheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme"
	informers "k8s.io/sample-controller/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	listers "k8s.io/sample-controller/pkg/generated/listers/samplecontroller/v1alpha1"
)

const controllerAgentName = "sample-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a ScaleSchedule is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a ScaleSchedule fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by ScaleSchedule"
	// MessageResourceSynced is the message used for an Event fired when a ScaleSchedule
	// is synced successfully
	MessageResourceSynced = "ScaleSchedule synced successfully"
)

// Controller is the controller implementation for ScaleSchedule resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	cronjobsLister batchlisters.CronJobLister
	cronjobsSynced cache.InformerSynced
	scaleschedulesLister        listers.ScaleScheduleLister
	scaleschedulesSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func convertToCronSyntax(timeStr string) (string) {
	// Split the time string into hours and minutes
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		fmt.Println("Invalid time format")
	}

	// Parse hours and minutes as integers
	hours, err := strconv.Atoi(timeParts[0])
	if err != nil || hours < 0 || hours > 23 {
		fmt.Println("Invalid hours")
	}

	minutes, err := strconv.Atoi(timeParts[1])
	if err != nil || minutes < 0 || minutes > 59 {
		fmt.Println("Invalid minutes")
	}

	// Construct cron syntax
	cronSyntax := fmt.Sprintf("%02d %02d * * *", minutes, hours)
	return cronSyntax
}

// NewController returns a new sample controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	cronjobInformer batchinformers.CronJobInformer,
	fooInformer informers.ScaleScheduleInformer) *Controller {
	logger := klog.FromContext(ctx)

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		cronjobsLister: cronjobInformer.Lister(),
		cronjobsSynced: cronjobInformer.Informer().HasSynced,
		scaleschedulesLister:        fooInformer.Lister(),
		scaleschedulesSynced:        fooInformer.Informer().HasSynced,
		workqueue:         workqueue.NewRateLimitingQueue(ratelimiter),
		recorder:          recorder,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when ScaleSchedule resources change
	fooInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueScaleSchedule,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueScaleSchedule(new)
		},
		DeleteFunc: controller.enqueueScaleSchedule,
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a ScaleSchedule resource then the handler will enqueue that ScaleSchedule resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	cronjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newCron := new.(*batchv1.CronJob)
			oldCron := old.(*batchv1.CronJob)
			if newCron.ResourceVersion == oldCron.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
		
	})

	return controller
}




// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	// Start the informer factories to begin populating the informer caches
	logger.Info("Starting ScaleSchedule controller")

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.cronjobsSynced, c.scaleschedulesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	// Launch two workers to process ScaleSchedule resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	logger := klog.FromContext(ctx)

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// ScaleSchedule resource to be synced.
		if err := c.syncHandler(ctx, key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		logger.Info("Successfully synced", "resourceName", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the ScaleSchedule resource
// with the current status of the resource.
func (c *Controller) syncHandler(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "resourceName", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the ScaleSchedule resource with this namespace/name
	foo, err := c.scaleschedulesLister.ScaleSchedules(namespace).Get(name)
	if err != nil {
		// The ScaleSchedule resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := foo.Spec.TargetRef.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}


	// Get the deployment with the name specified in ScaleSchedule.spec
for _, entry := range foo.Spec.Schedule {
	fmt.Printf("Schedule entry - At: %s, Replicas: %d\n", entry.At, entry.Replicas)

	cronJobName := fmt.Sprintf("%s-replicas-%d", deploymentName, entry.Replicas)
	cronjob, err := c.cronjobsLister.CronJobs(foo.Namespace).Get(cronJobName)

	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
			cronjob, err = c.kubeclientset.BatchV1().CronJobs(foo.Namespace).Create(context.TODO(), newCronJob(foo , entry.At, entry.Replicas), metav1.CreateOptions{})
				
	}



	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// // If the Deployment is not controlled by this ScaleSchedule resource, we should log
	// // a warning to the event recorder and return error msg.
	if !metav1.IsControlledBy(cronjob, foo) {
		msg := fmt.Sprintf(MessageResourceExists, cronjob.Name)
		c.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	// If this number of the replicas on the ScaleSchedule resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	// foo.Spec.Schedule != nil && 
	cronSyntax := convertToCronSyntax(foo.Spec.Schedule[0].At)
	if cronSyntax != cronjob.Spec.Schedule {
		logger.V(4).Info("Update cronjob Schedule", "currentSchedule", cronSyntax, "desiredSchedule", cronjob.Spec.Schedule)
		cronjob, err = c.kubeclientset.BatchV1().CronJobs(foo.Namespace).Update(context.TODO(), newCronJob(foo, entry.At, entry.Replicas), metav1.UpdateOptions{})
	}


	


	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// // Finally, we update the status block of the ScaleSchedule resource to reflect the
	// // current state of the world
	// err = c.updateScaleScheduleStatus(foo, deployment)
	// if err != nil {
	// 	return err
	// }

	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	}

	//cronJobName := fmt.Sprintf("%s-replicas-%d", foo.Spec.TargetRef.DeploymentName	, foo.Spec.Schedule[0].Replicas)
	//cronjob, err := c.cronjobsLister.CronJobs(foo.Namespace).Get()
	cronJobs, err := c.kubeclientset.BatchV1().CronJobs(foo.Namespace).List(context.TODO(), metav1.ListOptions{})

	// Iterate through each CronJob
	for _, cronJob := range cronJobs.Items {

		// Get the owner references of the CronJob
		ownerReferences := cronJob.GetOwnerReferences()
		// Iterate through owner references
		for _, ownerRef := range ownerReferences {
			// Print the Kind of the owner reference
			fmt.Printf("OwnerRef Kind: %s\n", ownerRef.Kind)
		
		
			var found bool
			for _, entry := range foo.Spec.Schedule {
				cronSyntax := convertToCronSyntax(entry.At)
				cronJobName := cronJob.Name
				re := regexp.MustCompile(`\d+$`)
				match := re.FindString(cronJobName)
				// Convert the matched string to an integer
				replica, err := strconv.Atoi(match)
				if err != nil {
					fmt.Errorf("Error converting string to integer:", err)
				}
				if cronSyntax == cronJob.Spec.Schedule &&  replica == entry.Replicas {
					found = true
					break
				}
			}
			// Check if the value was found
			if !found && ownerRef.Kind == "ScaleSchedule" {
				fmt.Printf("Value %s not found in the array\n", cronJob.Name)
				// Delete the CronJob if it doesn't exist in the CRD spec
				pp := metav1.DeletePropagationBackground
				err := c.kubeclientset.BatchV1().CronJobs(foo.Namespace).Delete(context.Background(), cronJob.Name, metav1.DeleteOptions{PropagationPolicy: &pp})
				if err != nil {
					fmt.Errorf("Failed to delete CronJob %s: %v\n", cronJob.Name, err)
				}
				fmt.Printf("Deleted CronJob %s\n", cronJob.Name)
				
			} else {
				fmt.Printf("Value %s found in the array\n", cronJob.Name)
			}
		}
	}
	return nil
}



// func (c *Controller) updateScaleScheduleStatus(foo *samplev1alpha1.ScaleSchedule, cronjob *batchv1.CronJob) error {
// 	// NEVER modify objects from the store. It's a read-only, local cache.
// 	// You can use DeepCopy() to make a deep copy of original object and modify this copy
// 	// Or create a copy manually for better performance


// 	fooCopy := foo.DeepCopy()
// 	fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
// 	// If the CustomResourceSubresources feature gate is not enabled,
// 	// we must use Update instead of UpdateStatus to update the Status block of the ScaleSchedule resource.
// 	// UpdateStatus will not allow changes to the Spec of the resource,
// 	// which is ideal for ensuring nothing other than resource status has been updated.
// 	_, err := c.sampleclientset.SamplecontrollerV1alpha1().ScaleSchedules(foo.Namespace).UpdateStatus(context.TODO(), fooCopy, metav1.UpdateOptions{})
// 	return err
// }

// enqueueScaleSchedule takes a ScaleSchedule resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than ScaleSchedule.
func (c *Controller) enqueueScaleSchedule(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}





// handleObject will take any resource implementing metav1.Object and attempt
// to find the ScaleSchedule resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that ScaleSchedule resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	logger := klog.FromContext(context.Background())
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		logger.V(4).Info("Recovered deleted object", "resourceName", object.GetName())
	}
	logger.V(4).Info("Processing object", "object", klog.KObj(object))
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a ScaleSchedule, we should not do anything more
		// with it.
		if ownerRef.Kind != "ScaleSchedule" {
			return
		}

		foo, err := c.scaleschedulesLister.ScaleSchedules(object.GetNamespace()).Get(ownerRef.Name)

		if err != nil {
			logger.V(4).Info("Ignore orphaned object", "object", klog.KObj(object), "foo", ownerRef.Name)
			return
		}



		c.enqueueScaleSchedule(foo)
		return
	}
}

// newDeployment creates a new Deployment for a ScaleSchedule resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the ScaleSchedule resource that 'owns' it.
func newCronJob(foo *samplev1alpha1.ScaleSchedule, at string, replicas int) *batchv1.CronJob {
// Loop through each ScheduleEntry in the spec.Schedule slice

	labels := map[string]string{
		"app":        "scale",
		"controller": foo.Name,
	}
	// Input string representing time in "HH:MM" format
	timeStr := at

	// Get the cron syntax
	cronSyntax := convertToCronSyntax(timeStr)

	scaleCommand := fmt.Sprintf("kubectl scale deployment %s --replicas=%d -n %s", foo.Spec.TargetRef.DeploymentName, replicas, foo.Spec.TargetRef.DeploymentNamespace)
	cronJobName := fmt.Sprintf("%s-replicas-%d", foo.Spec.TargetRef.DeploymentName, replicas)
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: cronJobName,
			Namespace: foo.Spec.TargetRef.DeploymentNamespace,
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, samplev1alpha1.SchemeGroupVersion.WithKind("ScaleSchedule")),
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: cronSyntax,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							ServiceAccountName: "scale-account",
							Containers: []corev1.Container{
								{
									Name: "kubectl-scaler",
									Image: "bitnami/kubectl:1.25.11",
									Command: []string{"/bin/sh", "-c", scaleCommand},
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,							
						},

					},

				},

			},
		},
	}
	
}
