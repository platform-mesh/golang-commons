package lifecycle

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/api"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/util"
	"github.com/platform-mesh/golang-commons/errors"
	"github.com/platform-mesh/golang-commons/logger"
	"github.com/platform-mesh/golang-commons/sentry"
)

func Reconcile(
	ctx context.Context,
	nName types.NamespacedName,
	instance runtimeobject.RuntimeObject,
	cl client.Client,
	l api.Lifecycle,
) (ctrl.Result, error) {
	ctx, span := otel.Tracer(l.Config().OperatorName).Start(ctx, fmt.Sprintf("%s.Reconcile", l.Config().ControllerName))
	defer span.End()

	result := ctrl.Result{}
	reconcileId := uuid.New().String()

	log := l.Log().
		MustChildLoggerWithAttributes("name", nName.Name, "namespace", nName.Namespace, "reconcile_id", reconcileId)
	sentryTags := sentry.Tags{"namespace": nName.Namespace, "name": nName.Name}

	ctx = logger.SetLoggerInContext(ctx, log)
	ctx = sentry.ContextWithSentryTags(ctx, sentryTags)

	log.Info().Msg("start reconcile")

	err := cl.Get(ctx, nName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info().Msg("instance not found. It was likely deleted")
			return ctrl.Result{}, nil
		}
		return HandleClientError("failed to retrieve instance", log, err, true, sentryTags)
	}

	originalCopy := instance.DeepCopyObject()
	inDeletion := instance.GetDeletionTimestamp() != nil
	generationChanged := true

	if l.Spreader() != nil && instance.GetDeletionTimestamp().IsZero() {
		reconcileRequired := l.Spreader().ReconcileRequired(instance, log)
		if !reconcileRequired {
			log.Info().Msg("skipping reconciliation, spread reconcile is active. No processing needed")
			return l.Spreader().OnNextReconcile(instance, log)
		}
	}

	// Manage Finalizers
	ferr := AddFinalizersIfNeeded(ctx, cl, instance, l.Subroutines(), l.Config().ReadOnly)
	if ferr != nil {
		return ctrl.Result{}, ferr
	}

	var condArr []v1.Condition
	if l.ConditionsManager() != nil {
		condArr = util.MustToInterface[api.RuntimeObjectConditions](instance, log).GetConditions()
		l.ConditionsManager().SetInstanceConditionUnknownIfNotSet(&condArr)
	}

	if l.PrepareContextFunc() != nil {
		localCtx, oErr := l.PrepareContextFunc()(ctx, instance)
		if oErr != nil {
			return HandleOperatorError(ctx, oErr, "failed to prepare context", generationChanged, l.Log())
		}
		ctx = localCtx
	}

	// In case of deletion execute the finalize subroutines in the reverse order as subroutine processing
	subroutines := make([]subroutine.BaseSubroutine, len(l.Subroutines()))
	copy(subroutines, l.Subroutines())
	if inDeletion {
		slices.Reverse(subroutines)
	}

	// Continue with reconciliation
	var hasNonRetryableError, stopChain bool
	for _, s := range subroutines {
		if l.ConditionsManager() != nil {
			l.ConditionsManager().SetSubroutineConditionToUnknownIfNotSet(&condArr, s, inDeletion, log)
		}

		// Set current condArr before reconciling the s
		if l.ConditionsManager() != nil {
			util.MustToInterface[api.RuntimeObjectConditions](instance, log).SetConditions(condArr)
		}

		var subResult subroutine.Result
		switch sub := s.(type) {
		case subroutine.ChainSubroutine:
			subResult = reconcileChainSubroutine(ctx, instance, sub, cl, l, log, generationChanged, sentryTags)
		case subroutine.Subroutine:
			subResult = reconcileSubroutine(ctx, instance, sub, cl, l, log, generationChanged, sentryTags)
		default:
			log.Error().Str("subroutine", s.GetName()).Msg("unknown subroutine type")
			continue
		}

		if l.ConditionsManager() != nil {
			condArr = util.MustToInterface[api.RuntimeObjectConditions](instance, log).GetConditions()
		}

		if subResult.Ctrl.RequeueAfter > 0 {
			if subResult.Ctrl.RequeueAfter < result.RequeueAfter || result.RequeueAfter == 0 {
				result.RequeueAfter = subResult.Ctrl.RequeueAfter
			}
		}

		switch subResult.Outcome {
		case subroutine.Continue:
			if l.ConditionsManager() != nil && subResult.Ctrl.RequeueAfter == 0 {
				l.ConditionsManager().SetSubroutineCondition(&condArr, s, subResult.Ctrl, nil, inDeletion, log)
			}
		case subroutine.StopChain:
			if l.ConditionsManager() != nil {
				l.ConditionsManager().SetSubroutineCondition(&condArr, s, subResult.Ctrl, nil, inDeletion, log)
			}

			log.Info().Str("subroutine", s.GetName()).Str("reason", subResult.Reason).Msg("stop chain requested")
			stopChain = true
		case subroutine.Skipped:
			if l.ConditionsManager() != nil {
				l.ConditionsManager().SetSubroutineCondition(&condArr, s, subResult.Ctrl, nil, inDeletion, log)
			}
		case subroutine.ErrorRetry:
			if l.ConditionsManager() != nil {
				l.ConditionsManager().
					SetSubroutineCondition(&condArr, s, subResult.Ctrl, subResult.Error, inDeletion, log)
				l.ConditionsManager().SetInstanceConditionReady(&condArr, v1.ConditionFalse)
				util.MustToInterface[api.RuntimeObjectConditions](instance, log).SetConditions(condArr)
			}

			if !l.Config().ReadOnly {
				if updateErr := updateStatus(ctx, cl, originalCopy, instance, log, generationChanged, sentryTags); updateErr != nil {
					log.Error().Err(updateErr).Msg("failed to update status during error retry")
				}
			}

			return subResult.Ctrl, subResult.Error
		case subroutine.ErrorContinue:
			hasNonRetryableError = true
			if l.ConditionsManager() != nil {
				l.ConditionsManager().
					SetSubroutineCondition(&condArr, s, subResult.Ctrl, subResult.Error, inDeletion, log)
			}

			log.Warn().Str("subroutine", s.GetName()).Err(subResult.Error).Msg("non-retryable error, continuing")
		case subroutine.ErrorStop:
			hasNonRetryableError = true
			if l.ConditionsManager() != nil {
				l.ConditionsManager().
					SetSubroutineCondition(&condArr, s, subResult.Ctrl, subResult.Error, inDeletion, log)
				l.ConditionsManager().SetInstanceConditionReady(&condArr, v1.ConditionFalse)
				util.MustToInterface[api.RuntimeObjectConditions](instance, log).SetConditions(condArr)
			}

			log.Warn().
				Str("subroutine", s.GetName()).
				Err(subResult.Error).
				Str("reason", subResult.Reason).
				Msg("non-retryable error, stopping chain")
			stopChain = true
		}

		if stopChain {
			break
		}
	}

	if result.RequeueAfter == 0 {
		if hasNonRetryableError {
			MarkResourceAsFinal(instance, log, condArr, v1.ConditionFalse, l)
		} else {
			MarkResourceAsFinal(instance, log, condArr, v1.ConditionTrue, l)
		}
	} else {
		if l.ConditionsManager() != nil {
			l.ConditionsManager().SetInstanceConditionReady(&condArr, v1.ConditionFalse)
		}
	}

	if l.ConditionsManager() != nil {
		util.MustToInterface[api.RuntimeObjectConditions](instance, log).SetConditions(condArr)
	}

	if !l.Config().ReadOnly {
		err = updateStatus(ctx, cl, originalCopy, instance, log, generationChanged, sentryTags)
		if err != nil {
			return result, err
		}
	}

	if l.Spreader() != nil && instance.GetDeletionTimestamp().IsZero() {
		original := instance.DeepCopyObject().(client.Object)
		removed := l.Spreader().RemoveRefreshLabelIfExists(instance)
		if removed {
			updateErr := cl.Patch(ctx, instance, client.MergeFrom(original))
			if updateErr != nil {
				return HandleClientError("failed to update instance", log, err, generationChanged, sentryTags)
			}
		}
	}

	log.Info().Msg("end reconcile")
	return result, nil
}

func reconcileSubroutine(
	ctx context.Context,
	instance runtimeobject.RuntimeObject,
	s subroutine.Subroutine,
	cl client.Client,
	l api.Lifecycle,
	log *logger.Logger,
	generationChanged bool,
	sentryTags map[string]string,
) subroutine.Result {
	subroutineLogger := log.ChildLogger("subroutine", s.GetName())
	ctx = logger.SetLoggerInContext(ctx, subroutineLogger)
	subroutineLogger.Debug().Msg("start subroutine")

	ctx, span := otel.Tracer(l.Config().OperatorName).
		Start(ctx, fmt.Sprintf("%s.reconcileSubroutine.%s", l.Config().ControllerName, s.GetName()))
	defer span.End()

	var ctrlResult ctrl.Result
	var err errors.OperatorError
	if instance.GetDeletionTimestamp() != nil {
		if containsFinalizer(instance, s.Finalizers(instance)) {
			subroutineLogger.Debug().Msg("finalizing instance")
			ctrlResult, err = s.Finalize(ctx, instance)
			subroutineLogger.Debug().Any("result", ctrlResult).Msg("finalized instance")
			if err == nil {
				// Remove finalizers unless requeue is requested
				if ferr := removeSubroutineFinalizerIfNeeded(
					ctx,
					instance,
					s,
					ctrlResult,
					l.Config().ReadOnly,
					cl,
				); ferr != nil {
					return subroutine.Retry(ferr)
				}
			}
		} else {
			return subroutine.Skip("no finalizer")
		}
	} else {
		subroutineLogger.Debug().Msg("processing instance")
		ctrlResult, err = s.Process(ctx, instance)
		subroutineLogger.Debug().Any("result", ctrlResult).Msg("processed instance")
	}

	result := subroutine.ResultFromSubroutine(ctrlResult, err)

	if result.IsError() && result.Sentry && generationChanged {
		sentry.CaptureError(result.Error, sentryTags)
	}

	if result.IsError() {
		subroutineLogger.Error().
			Err(result.Error).
			Bool("retry", result.Outcome == subroutine.ErrorRetry).
			Msg("subroutine ended with error")
	}

	subroutineLogger.Debug().Msg("end subroutine")
	return result
}

func reconcileChainSubroutine(
	ctx context.Context,
	instance runtimeobject.RuntimeObject,
	s subroutine.ChainSubroutine,
	cl client.Client,
	l api.Lifecycle,
	log *logger.Logger,
	generationChanged bool,
	sentryTags map[string]string,
) subroutine.Result {
	subroutineLogger := log.ChildLogger("subroutine", s.GetName())
	ctx = logger.SetLoggerInContext(ctx, subroutineLogger)

	ctx, span := otel.Tracer(l.Config().OperatorName).
		Start(ctx, fmt.Sprintf("%s.reconcileSubroutine.%s", l.Config().ControllerName, s.GetName()))
	defer span.End()

	var result subroutine.Result
	if instance.GetDeletionTimestamp() != nil {
		if containsFinalizer(instance, s.Finalizers(instance)) {
			result = s.Finalize(ctx, instance)
			if !result.IsError() {
				if ferr := removeChainSubroutineFinalizerIfNeeded(
					ctx,
					instance,
					s,
					result.Ctrl,
					l.Config().ReadOnly,
					cl,
				); ferr != nil {
					return subroutine.Retry(ferr)
				}
			}
		} else {
			return subroutine.Skip("no finalizer")
		}
	} else {
		result = s.Process(ctx, instance)
	}

	if result.IsError() && result.Sentry && generationChanged {
		sentry.CaptureError(result.Error, sentryTags)
	}

	if result.IsError() {
		subroutineLogger.Error().Err(result.Error).Msg("subroutine error")
	}

	return result
}

func containsFinalizer(o client.Object, subroutineFinalizers []string) bool {
	for _, subroutineFinalizer := range subroutineFinalizers {
		if controllerutil.ContainsFinalizer(o, subroutineFinalizer) {
			return true
		}
	}
	return false
}

func removeSubroutineFinalizerIfNeeded(
	ctx context.Context,
	instance runtimeobject.RuntimeObject,
	s subroutine.Subroutine,
	result ctrl.Result,
	readonly bool,
	cl client.Client,
) error {
	if readonly {
		return nil
	}

	if result.RequeueAfter == 0 {
		update := false
		original := instance.DeepCopyObject().(client.Object)
		for _, f := range s.Finalizers(instance) {
			needsUpdate := controllerutil.RemoveFinalizer(instance, f)
			if needsUpdate {
				update = true
			}
		}
		if update {
			err := cl.Patch(ctx, instance, client.MergeFrom(original))
			if err != nil {
				return fmt.Errorf("failed to update instance: %w", err)
			}
		}
	}

	return nil
}

func removeChainSubroutineFinalizerIfNeeded(
	ctx context.Context,
	instance runtimeobject.RuntimeObject,
	s subroutine.ChainSubroutine,
	result ctrl.Result,
	readonly bool,
	cl client.Client,
) error {
	if readonly {
		return nil
	}

	if result.RequeueAfter == 0 {
		update := false
		original := instance.DeepCopyObject().(client.Object)
		for _, f := range s.Finalizers(instance) {
			needsUpdate := controllerutil.RemoveFinalizer(instance, f)
			if needsUpdate {
				update = true
			}
		}
		if update {
			err := cl.Patch(ctx, instance, client.MergeFrom(original))
			if err != nil {
				return fmt.Errorf("failed to update instance: %w", err)
			}
		}
	}

	return nil
}

func updateStatus(
	ctx context.Context,
	cl client.Client,
	original runtime.Object,
	current runtimeobject.RuntimeObject,
	log *logger.Logger,
	generationChanged bool,
	sentryTags sentry.Tags,
) error {
	currentUn, err := runtime.DefaultUnstructuredConverter.ToUnstructured(current)
	if err != nil {
		return err
	}

	originalUn, err := runtime.DefaultUnstructuredConverter.ToUnstructured(original)
	if err != nil {
		return err
	}

	currentStatus, _, err := unstructured.NestedFieldCopy(currentUn, "status")
	if err != nil {
		return err
	}

	originalStatus, _, err := unstructured.NestedFieldCopy(originalUn, "status")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(currentStatus, originalStatus) {
		log.Info().Msg("skipping status update, since they are equal")
		return nil
	}

	log.Info().Msg("updating resource status")
	err = cl.Status().Update(ctx, current)
	if err != nil {
		if kerrors.IsConflict(err) {
			log.Warn().Err(err).Msg("cannot update reconciliation Conditions, kubernetes client error")
		} else {
			log.Error().Err(err).Msg("cannot update status, kubernetes client error")
			if generationChanged {
				sentry.CaptureError(err, sentryTags, sentry.Extras{"message": "Updating of instance status failed"})
			}
		}
		return err
	}

	return nil
}

func HandleClientError(
	msg string,
	log *logger.Logger,
	err error,
	generationChanged bool,
	sentryTags sentry.Tags,
) (ctrl.Result, error) {
	log.Error().Err(err).Msg(msg)
	if generationChanged {
		sentry.CaptureError(err, sentryTags)
	}

	return ctrl.Result{}, err
}

func MarkResourceAsFinal(
	instance runtimeobject.RuntimeObject,
	log *logger.Logger,
	conditions []v1.Condition,
	status v1.ConditionStatus,
	l api.Lifecycle,
) {
	if l.Spreader() != nil && instance.GetDeletionTimestamp().IsZero() {
		instanceStatusObj := util.MustToInterface[api.RuntimeObjectSpreadReconcileStatus](instance, log)
		l.Spreader().SetNextReconcileTime(instanceStatusObj, log)
		l.Spreader().UpdateObservedGeneration(instanceStatusObj, log)
	}

	if l.ConditionsManager() != nil {
		l.ConditionsManager().SetInstanceConditionReady(&conditions, status)
	}
}

func AddFinalizersIfNeeded(
	ctx context.Context,
	cl client.Client,
	instance runtimeobject.RuntimeObject,
	subroutines []subroutine.BaseSubroutine,
	readonly bool,
) error {
	if readonly {
		return nil
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		return nil
	}

	update := false
	original := instance.DeepCopyObject().(client.Object)
	for _, s := range subroutines {
		if len(s.Finalizers(instance)) > 0 {
			needsUpdate := addFinalizerIfNeeded(instance, s)
			if needsUpdate {
				update = true
			}
		}
	}
	if update {
		err := cl.Patch(ctx, instance, client.MergeFrom(original))
		if err != nil {
			return err
		}
	}
	return nil
}

func addFinalizerIfNeeded(instance runtimeobject.RuntimeObject, s subroutine.BaseSubroutine) bool {
	update := false
	for _, f := range s.Finalizers(instance) {
		needsUpdate := controllerutil.AddFinalizer(instance, f)
		if needsUpdate {
			update = true
		}
	}
	return update
}

func HandleOperatorError(
	ctx context.Context,
	operatorError errors.OperatorError,
	msg string,
	generationChanged bool,
	log *logger.Logger,
) (ctrl.Result, error) {
	log.Error().
		Bool("retry", operatorError.Retry()).
		Bool("sentry", operatorError.Sentry()).
		Err(operatorError.Err()).
		Msg(msg)
	if generationChanged && operatorError.Sentry() {
		sentry.CaptureError(operatorError.Err(), sentry.GetSentryTagsFromContext(ctx))
	}

	if operatorError.Retry() {
		return ctrl.Result{}, operatorError.Err()
	}

	return ctrl.Result{}, nil
}

func ValidateInterfaces(instance runtimeobject.RuntimeObject, log *logger.Logger, l api.Lifecycle) error {
	if l.Spreader() != nil {
		_, err := util.ToInterface[api.RuntimeObjectSpreadReconcileStatus](instance, log)
		if err != nil {
			return err
		}
	}
	if l.ConditionsManager() != nil {
		_, err := util.ToInterface[api.RuntimeObjectConditions](instance, log)
		if err != nil {
			return err
		}
	}
	return nil
}
