/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme"
)

// ScaleSchedulesGetter has a method to return a ScaleScheduleInterface.
// A group's client should implement this interface.
type ScaleSchedulesGetter interface {
	ScaleSchedules(namespace string) ScaleScheduleInterface
}

// ScaleScheduleInterface has methods to work with ScaleSchedule resources.
type ScaleScheduleInterface interface {
	Create(ctx context.Context, scaleSchedule *v1alpha1.ScaleSchedule, opts v1.CreateOptions) (*v1alpha1.ScaleSchedule, error)
	Update(ctx context.Context, scaleSchedule *v1alpha1.ScaleSchedule, opts v1.UpdateOptions) (*v1alpha1.ScaleSchedule, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ScaleSchedule, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ScaleScheduleList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ScaleSchedule, err error)
	ScaleScheduleExpansion
}

// scaleSchedules implements ScaleScheduleInterface
type scaleSchedules struct {
	client rest.Interface
	ns     string
}

// newScaleSchedules returns a ScaleSchedules
func newScaleSchedules(c *SamplecontrollerV1alpha1Client, namespace string) *scaleSchedules {
	return &scaleSchedules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the scaleSchedule, and returns the corresponding scaleSchedule object, and an error if there is any.
func (c *scaleSchedules) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ScaleSchedule, err error) {
	result = &v1alpha1.ScaleSchedule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scaleschedules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ScaleSchedules that match those selectors.
func (c *scaleSchedules) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ScaleScheduleList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ScaleScheduleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scaleschedules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested scaleSchedules.
func (c *scaleSchedules) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("scaleschedules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a scaleSchedule and creates it.  Returns the server's representation of the scaleSchedule, and an error, if there is any.
func (c *scaleSchedules) Create(ctx context.Context, scaleSchedule *v1alpha1.ScaleSchedule, opts v1.CreateOptions) (result *v1alpha1.ScaleSchedule, err error) {
	result = &v1alpha1.ScaleSchedule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("scaleschedules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(scaleSchedule).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a scaleSchedule and updates it. Returns the server's representation of the scaleSchedule, and an error, if there is any.
func (c *scaleSchedules) Update(ctx context.Context, scaleSchedule *v1alpha1.ScaleSchedule, opts v1.UpdateOptions) (result *v1alpha1.ScaleSchedule, err error) {
	result = &v1alpha1.ScaleSchedule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scaleschedules").
		Name(scaleSchedule.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(scaleSchedule).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the scaleSchedule and deletes it. Returns an error if one occurs.
func (c *scaleSchedules) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scaleschedules").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *scaleSchedules) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scaleschedules").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched scaleSchedule.
func (c *scaleSchedules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ScaleSchedule, err error) {
	result = &v1alpha1.ScaleSchedule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("scaleschedules").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}