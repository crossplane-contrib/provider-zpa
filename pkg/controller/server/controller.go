/*
Copyright 2021 The Crossplane Authors.
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

package server

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	zpa "github.com/haarchri/zpa-go-client/pkg/client"
	"github.com/haarchri/zpa-go-client/pkg/client/app_server_controller"
	"github.com/haarchri/zpa-go-client/pkg/models"

	v1alpha1 "github.com/crossplane-contrib/provider-zpa/apis/server/v1alpha1"
	zpaclient "github.com/crossplane-contrib/provider-zpa/pkg/client"
)

const (
	errNotServer      = "managed resource is not an Server custom resource"
	errCreateFailed   = "cannot create Server"
	errUpdateFailed   = "connot update Server"
	errDescribeFailed = "cannot describe Server"
	errDeleteFailed   = "cannot delete Server"
)

// SetupServer adds a controller that reconciles Servers.
func SetupServer(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.ServerKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.Server{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ServerGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: zpa.New}),
			managed.WithInitializers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(transport runtime.ClientTransport, formats strfmt.Registry) *zpa.ZscalerPrivateAccessAPIPortal
}

type external struct {
	client *zpa.ZscalerPrivateAccessAPIPortal
	kube   client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Server)
	if !ok {
		return nil, errors.New(errNotServer)
	}

	cfg, err := zpaclient.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}

	client := c.newClientFn(cfg, strfmt.Default)
	return &external{client, c.kube}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Server)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotServer)
	}

	id := meta.GetExternalName(cr)
	if id == "" {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &app_server_controller.GetAppServerUsingGET1Params{
		Context:    ctx,
		ServerID:   id,
		CustomerID: customerID,
	}
	resp, reqErr := e.client.AppServerController.GetAppServerUsingGET1(req)
	if reqErr != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(IsNotFound, reqErr), errDescribeFailed)
	}

	cr.Status.AtProvider = generateObservation(resp)

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	e.LateInitialize(cr, resp)

	cr.Status.SetConditions(v1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate(&cr.Spec.ForProvider, resp),
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Server)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotServer)
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &app_server_controller.AddAppServerUsingPOST1Params{
		Context:    ctx,
		CustomerID: customerID,
		Server: &models.ApplicationServer{
			Name:              cr.Spec.ForProvider.Name,
			Address:           cr.Spec.ForProvider.Address,
			ConfigSpace:       cr.Spec.ForProvider.ConfigSpace,
			Description:       cr.Spec.ForProvider.Description,
			AppServerGroupIds: cr.Spec.ForProvider.ServerGroups,
			Enabled:           zpaclient.BoolValue(cr.Spec.ForProvider.Enabled),
		},
	}

	resp, err := e.client.AppServerController.AddAppServerUsingPOST1(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, resp.Payload.ID)
	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil

}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Server)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotServer)
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &app_server_controller.UpdateAppServerUsingPUT1Params{
		Context:    ctx,
		CustomerID: customerID,
		ServerID:   meta.GetExternalName(cr),
		Server: &models.ApplicationServer{
			Name:        cr.Spec.ForProvider.Name,
			Address:     cr.Spec.ForProvider.Address,
			ConfigSpace: cr.Spec.ForProvider.ConfigSpace,
			Description: cr.Spec.ForProvider.Description,
			Enabled:     zpaclient.BoolValue(cr.Spec.ForProvider.Enabled),
		},
	}

	if len(cr.Spec.ForProvider.ServerGroups) > 0 {
		req.Server.AppServerGroupIds = cr.Spec.ForProvider.ServerGroups
	} else {
		req.Server.AppServerGroupIds = make([]string, 0)
	}

	if _, _, err := e.client.AppServerController.UpdateAppServerUsingPUT1(req); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Server)
	if !ok {
		return errors.New(errNotServer)
	}

	id := meta.GetExternalName(cr)
	if id == "" {
		return errors.New(errNotServer)
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	// Remove the reference to this server from server groups.
	if len(cr.Spec.ForProvider.ServerGroups) != 0 {
		upreq := &app_server_controller.UpdateAppServerUsingPUT1Params{
			Context:    ctx,
			CustomerID: customerID,
			ServerID:   meta.GetExternalName(cr),
			Server: &models.ApplicationServer{
				Name:              cr.Spec.ForProvider.Name,
				Address:           cr.Spec.ForProvider.Address,
				ConfigSpace:       cr.Spec.ForProvider.ConfigSpace,
				Description:       cr.Spec.ForProvider.Description,
				AppServerGroupIds: make([]string, 0),
				Enabled:           zpaclient.BoolValue(cr.Spec.ForProvider.Enabled),
			},
		}

		if _, _, err := e.client.AppServerController.UpdateAppServerUsingPUT1(upreq); err != nil {
			return errors.Wrap(err, errUpdateFailed)
		}
	}

	// Delete Server finally
	req := &app_server_controller.DeleteAppServerUsingDELETE1Params{
		Context:    ctx,
		ServerID:   id,
		CustomerID: customerID,
	}

	_, err := e.client.AppServerController.DeleteAppServerUsingDELETE1(req)
	if err != nil {
		return errors.Wrap(err, errDeleteFailed)
	}

	return nil
}

func (e *external) LateInitialize(cr *v1alpha1.Server, obj *app_server_controller.GetAppServerUsingGET1OK) { // nolint:gocyclo

	if cr.Spec.ForProvider.ConfigSpace == "" {
		cr.Spec.ForProvider.ConfigSpace = obj.Payload.ConfigSpace
	}

}

// generateObservation generates observation for the input object app_server_controller.GetAppServerUsingGET1OK
func generateObservation(in *app_server_controller.GetAppServerUsingGET1OK) v1alpha1.Observation {
	cr := v1alpha1.Observation{}

	obj := in.Payload

	cr.CreationTime = obj.CreationTime
	cr.ID = obj.ID
	cr.ModifiedBy = obj.ModifiedBy
	cr.ModifiedTime = obj.ModifiedTime

	return cr
}

// isUpToDate checks whether there is a change in any of the modifiable fields.
func isUpToDate(cr *v1alpha1.ServerParameters, gobj *app_server_controller.GetAppServerUsingGET1OK) bool { // nolint:gocyclo
	obj := gobj.Payload

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.Description), zpaclient.StringToPtr(obj.Description)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.ConfigSpace), zpaclient.StringToPtr(obj.ConfigSpace)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.Address), zpaclient.StringToPtr(obj.Address)) {
		return false
	}

	if !zpaclient.IsEqualStringArrayContent(cr.ServerGroups, obj.AppServerGroupIds) {
		return false
	}

	if !zpaclient.IsEqualString(cr.Name, obj.Name) {
		return false
	}

	return true
}
