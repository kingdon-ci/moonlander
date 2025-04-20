package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"

	moonsv1alpha1 "github.com/kingdon-ci/moonlander/api/v1alpha1"
)

type LandingReconciler struct {
	client.Client
}

// 	// Basic stub logic
// 	return reconcile.Result{}, nil
// }

func (r *LandingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
// func (r *LandingReconciler) Reconcile(ctx context.Context, req client.ObjectKey) error {
	logger := log.FromContext(ctx)

	// 1. Fetch the Landing resource
	var landing moonsv1alpha1.Landing
	if err := r.Get(ctx, req.NamespacedName, &landing); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Landing resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Failed to get Landing")
		return reconcile.Result{}, err
	}

	// 2. Resolve child kubeconfig secret name
	childSecretName := fmt.Sprintf("kubernetes-%s-admin-kubeconfig", landing.Name)
	if landing.Spec.KubeconfigSecretName != "" {
		childSecretName = landing.Spec.KubeconfigSecretName
	}

	childNamespace := landing.Namespace
	if landing.Spec.KubeconfigSecretNamespace != "" {
		childNamespace = landing.Spec.KubeconfigSecretNamespace
	}

	// 3. Get the child kubeconfig secret
	var childSecret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Name: childSecretName, Namespace: childNamespace}, &childSecret); err != nil {
		logger.Error(err, "Failed to get child kubeconfig secret", "secret", childSecretName)
		return reconcile.Result{}, err
	}

	kubeconfigBytes, ok := childSecret.Data["admin.conf"]
	if !ok {
		err := fmt.Errorf("child kubeconfig secret %s missing 'admin.conf' key", childSecretName)
		logger.Error(err, "Malformed secret")
		return reconcile.Result{}, err
	}

	logger.Info("Successfully fetched child kubeconfig secret")

	// 4. Read parent kubeconfig from mounted file
	parentKubeconfigPath := "/etc/moonlander/parent-kubeconfig"
	parentKubeconfigBytes, err := os.ReadFile(parentKubeconfigPath)
	if err != nil {
		err := fmt.Errorf("failed to read parent kubeconfig from %s: %w", parentKubeconfigPath, err)
		logger.Error(err, "Could not load parent kubeconfig")
		return reconcile.Result{}, err
	}
	logger.Info("Successfully loaded parent kubeconfig")

	// 5. Build REST config from parent kubeconfig
	parentRestConfig, err := clientcmd.RESTConfigFromKubeConfig(parentKubeconfigBytes)
	if err != nil {
		logger.Error(err, "Failed to build REST config from parent kubeconfig")
		return reconcile.Result{}, err
	}

	// 6. Build REST config from child kubeconfig
	childRestConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	if err != nil {
		logger.Error(err, "Failed to build REST config from child kubeconfig")
		return reconcile.Result{}, err
	}

	// 7. Create Kubernetes clients for parent and child
	_, err = kubernetes.NewForConfig(parentRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create client for parent cluster")
		return reconcile.Result{}, err
	}

	childClient, err := client.New(childRestConfig, client.Options{})
	if err != nil {
		logger.Error(err, "Failed to create client for child cluster")
		return reconcile.Result{}, err
	}

	logger.Info("Successfully created Kubernetes clients for parent and child clusters")

	// 8. Determine secret name and namespace for writing parent kubeconfig into child cluster
	writeSecretName := "remote-kubeconfig"
	if landing.Spec.WriteKubeconfigSecretName != "" {
		writeSecretName = landing.Spec.WriteKubeconfigSecretName
	}
	targetNamespace := landing.Namespace
	if landing.Spec.TargetNamespace != "" {
		targetNamespace = landing.Spec.TargetNamespace
	}

	// 9. Encode the parent kubeconfig
	encodedParentKubeconfig := base64.StdEncoding.EncodeToString(parentKubeconfigBytes)

	// 10. Prepare the Secret object
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      writeSecretName,
			Namespace: targetNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"value": []byte(encodedParentKubeconfig),
		},
	}

	// 11. Create or Update the secret in the child cluster
	existingSecret := &corev1.Secret{}
	err = childClient.Get(ctx, client.ObjectKey{Name: writeSecretName, Namespace: targetNamespace}, existingSecret)
	if err != nil && errors.IsNotFound(err) {
		if err := childClient.Create(ctx, newSecret); err != nil {
			logger.Error(err, "Failed to create kubeconfig secret in child cluster")
			return reconcile.Result{}, err
		}
		logger.Info("Successfully created parent kubeconfig secret in child cluster")
	} else if err == nil {
		existingSecret.Data = newSecret.Data
		if err := childClient.Update(ctx, existingSecret); err != nil {
			logger.Error(err, "Failed to update kubeconfig secret in child cluster")
			return reconcile.Result{}, err
		}
		logger.Info("Successfully updated parent kubeconfig secret in child cluster")
	} else {
		logger.Error(err, "Failed to get existing secret in child cluster")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *LandingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&moonsv1alpha1.Landing{}).
		Complete(r)
}
