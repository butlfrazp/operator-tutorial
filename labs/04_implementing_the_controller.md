# Implementing the Controller

In `songs-operator/controllers/songsconfig_controller.go` there are two functions `Reconcile` and `SetupWithManager`.
The majority of the heavy lift will be done in the `Reconcile` method.
This method handles all the business logic that needs to happend with a SongsConfig CRD is applied or deleted.
There are some annotations above the function that start with `+kubebuilder:rbac`.
These annotations specify permissions and create RBAC manifests.
More on RBAC can be found [here](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

## Setting up the Reconcile Function

There are two things we need to handle in Reconcile.

1) Creating a new deployment if the referenced CRD does not exist.
1) Update an existing deployment if the referenced CRD exists.

### Creating new deployment if CRD does not exist

```golang
log := log.FromContext(ctx)

songsConfigs := &configv1alpha1.SongsConfig{}
err := r.Get(ctx, req.NamespacedName, songsConfigs)
if err != nil {
    if errors.IsNotFound(err) {
        log.Info("SongConfig resource not found. Ignoring since object must be deleted")
        return ctrl.Result{}, nil
    }
    log.Error(err, "Failed to get SongConfig")
    return ctrl.Result{}, err
}
```

```golang
found := &appsv1.Deployment{}
err = r.Get(ctx, types.NamespacedName{Name: songsConfigs.Name, Namespace: songsConfigs.Namespace}, found)
if err != nil {
    if errors.IsNotFound(err) {
        dep, err := r.deploymentForSongService(songsConfigs)
        if err != nil {
            log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
            return ctrl.Result{}, err
        }

        log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
        err = r.Create(ctx, dep)
        if err != nil {
            log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }
}
```

### Updating an Existing Deployment

```golang
songs := &songsConfigs.Spec.Songs
b := []byte(found.Spec.Template.Spec.Containers[0].Env[0].Value)
deployedSongs := &[]configv1alpha1.Song{}

err = json.Unmarshal(b, deployedSongs)
if err != nil {
    return ctrl.Result{}, err
}

if !reflect.DeepEqual(songs, deployedSongs) {
    log.Info("Updating the pod")
    bytes, err := json.Marshal(songs)
    if err != nil {
        return ctrl.Result{}, err
    }
    found.Spec.Template.Spec.Containers[0].Env[0].Value = string(bytes)

    err = r.Update(ctx, found)
    if err != nil {
        log.Error(err, "Failed updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
        return ctrl.Result{}, err
    }

    return ctrl.Result{RequeueAfter: time.Minute}, nil
}
return ctrl.Result{}, nil
```

## Setting up the SetupWithManager Function

```golang
func (r *SongsConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.SongsConfig{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
```