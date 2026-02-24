package conditions

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	"github.com/platform-mesh/golang-commons/logger"
)

const (
	ConditionReady = "Ready"

	messageResourceReady      = "The resource is ready"
	messageResourceNotReady   = "The resource is not ready"
	messageResourceProcessing = "The resource is processing"

	reasonComplete   = "Complete"
	reasonProcessing = "Processing"
	reasonError      = "Error"

	// additional Reason values for chain functionality
	reasonSkipped       = "Skipped"
	reasonStopChain     = "StopChain"
	reasonErrorRetry    = "ErrorRetry"
	reasonErrorContinue = "ErrorContinue"
	reasonErrorStop     = "ErrorStop"

	subroutineReadyConditionFormatString    = "%s_Ready"
	subroutineFinalizeConditionFormatString = "%s_Finalize"

	subroutineMessageProcessingFormatString = "The %s is processing"
	subroutineMessageCompleteFormatString   = "The %s is complete"
	subroutineMessageErrorFormatString      = "The %s has an error: %s"

	// additional Message values for chain functionality
	subroutineMessageSkippedFormatString         = "The %s was skipped"
	subroutineMessageSkippedReasonFormatString   = "The %s was skipped: %s"
	subroutineMessageStopChainFormatString       = "The %s stopped chain"
	subroutineMessageStopChainReasonFormatString = "The %s stopped chain: %s"
	subroutineMessageErrorRetryFormatString      = "The %s has a retryable error: %s"
	subroutineMessageErrorContinueFormatString   = "The %s has a non-blocking error: %s"
	subroutineMessageErrorStopFormatString       = "The %s has a terminal error: %s"
)

type ConditionManager struct{}

func NewConditionManager() *ConditionManager {
	return &ConditionManager{}
}

// Set the Condition of the instance to be ready
func (c *ConditionManager) SetInstanceConditionReady(
	conditions *[]metav1.Condition,
	status metav1.ConditionStatus,
) bool {
	var msg string
	switch status {
	case metav1.ConditionTrue:
		msg = messageResourceReady
	case metav1.ConditionFalse:
		msg = messageResourceNotReady
	default:
		msg = messageResourceProcessing
	}
	return meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  status,
		Message: msg,
		Reason:  reasonComplete,
	})
}

// Set the Condition to be Unknown in case it is not set yet
func (c *ConditionManager) SetInstanceConditionUnknownIfNotSet(conditions *[]metav1.Condition) bool {
	existingCondition := meta.FindStatusCondition(*conditions, ConditionReady)
	if existingCondition == nil {
		return c.SetInstanceConditionReady(conditions, metav1.ConditionUnknown)
	}
	return false
}

func (c *ConditionManager) SetSubroutineConditionToUnknownIfNotSet(
	conditions *[]metav1.Condition,
	subroutine subroutine.BaseSubroutine,
	isFinalize bool,
	log *logger.Logger,
) bool {
	conditionName, conditionMessage := getConditionNameAndMessage(subroutine, isFinalize)

	existingCondition := meta.FindStatusCondition(*conditions, conditionName)
	if existingCondition == nil {
		changed := meta.SetStatusCondition(
			conditions,
			metav1.Condition{
				Type:    conditionName,
				Status:  metav1.ConditionUnknown,
				Message: fmt.Sprintf(subroutineMessageProcessingFormatString, conditionMessage),
				Reason:  reasonProcessing,
			},
		)
		if changed {
			log.Info().Str("type", conditionName).Msg("updated condition")
		}
		return changed
	}
	return false
}

func getConditionNameAndMessage(subroutine subroutine.BaseSubroutine, isFinalize bool) (string, string) {
	conditionName := fmt.Sprintf(subroutineReadyConditionFormatString, subroutine.GetName())
	conditionMessage := "subroutine"
	if isFinalize {
		conditionName = fmt.Sprintf(subroutineFinalizeConditionFormatString, subroutine.GetName())
		conditionMessage = "subroutine finalization"
	}
	return conditionName, conditionMessage
}

// Set Subroutines Conditions
func (c *ConditionManager) SetSubroutineCondition(
	conditions *[]metav1.Condition,
	subroutine subroutine.BaseSubroutine,
	subroutineResult ctrl.Result,
	subroutineErr error,
	isFinalize bool,
	log *logger.Logger,
) bool {
	conditionName, conditionMessage := getConditionNameAndMessage(subroutine, isFinalize)

	// processing complete
	if subroutineErr == nil && subroutineResult.RequeueAfter == 0 {
		return meta.SetStatusCondition(
			conditions,
			metav1.Condition{
				Type:    conditionName,
				Status:  metav1.ConditionTrue,
				Message: fmt.Sprintf(subroutineMessageCompleteFormatString, conditionMessage),
				Reason:  reasonComplete,
			},
		)
	}
	// processing is still processing
	if subroutineErr == nil && subroutineResult.RequeueAfter > 0 {
		return meta.SetStatusCondition(
			conditions,
			metav1.Condition{
				Type:    conditionName,
				Status:  metav1.ConditionUnknown,
				Message: fmt.Sprintf(subroutineMessageProcessingFormatString, conditionMessage),
				Reason:  reasonProcessing,
			},
		)
	}
	// processing failed
	var sErr error
	if subroutineErr != nil {
		sErr = subroutineErr
	}
	changed := meta.SetStatusCondition(
		conditions,
		metav1.Condition{
			Type:    conditionName,
			Status:  metav1.ConditionFalse,
			Message: fmt.Sprintf(subroutineMessageErrorFormatString, conditionMessage, sErr),
			Reason:  reasonError,
		},
	)
	if changed {
		log.Info().Str("type", conditionName).Msg("updated condition")
	}
	return changed
}

func (c *ConditionManager) SetSubroutineConditionFromResult(
	conditions *[]metav1.Condition,
	sub subroutine.BaseSubroutine,
	result subroutine.Result,
	isFinalize bool,
	log *logger.Logger,
) bool {
	conditionName, conditionMessage := getConditionNameAndMessage(sub, isFinalize)

	condition := metav1.Condition{Type: conditionName}

	switch result.Outcome {
	case subroutine.Continue:
		condition.Status = metav1.ConditionTrue
		condition.Reason = reasonComplete
		condition.Message = fmt.Sprintf(subroutineMessageCompleteFormatString, conditionMessage)
		if result.Ctrl.RequeueAfter > 0 {
			condition.Status = metav1.ConditionUnknown
			condition.Reason = reasonProcessing
			condition.Message = fmt.Sprintf(subroutineMessageProcessingFormatString, conditionMessage)
		}
	case subroutine.Skipped:
		condition.Status = metav1.ConditionTrue
		condition.Reason = reasonSkipped
		condition.Message = formatMessageWithOptionalReason(
			subroutineMessageSkippedFormatString, subroutineMessageSkippedReasonFormatString,
			conditionMessage, result.Reason,
		)
	case subroutine.StopChain:
		condition.Status = metav1.ConditionTrue
		condition.Reason = reasonStopChain
		condition.Message = formatMessageWithOptionalReason(
			subroutineMessageStopChainFormatString, subroutineMessageStopChainReasonFormatString,
			conditionMessage, result.Reason,
		)
	case subroutine.ErrorRetry:
		condition.Status = metav1.ConditionFalse
		condition.Reason = reasonErrorRetry
		condition.Message = fmt.Sprintf(subroutineMessageErrorRetryFormatString, conditionMessage, result.Error)
	case subroutine.ErrorContinue:
		condition.Status = metav1.ConditionFalse
		condition.Reason = reasonErrorContinue
		condition.Message = fmt.Sprintf(subroutineMessageErrorContinueFormatString, conditionMessage, result.Error)
	case subroutine.ErrorStop:
		condition.Status = metav1.ConditionFalse
		condition.Reason = reasonErrorStop
		condition.Message = fmt.Sprintf(subroutineMessageErrorStopFormatString, conditionMessage, result.Error)
	default:
		condition.Status = metav1.ConditionUnknown
		condition.Reason = reasonProcessing
		condition.Message = fmt.Sprintf(subroutineMessageProcessingFormatString, conditionMessage)
	}

	changed := meta.SetStatusCondition(conditions, condition)
	if changed {
		log.Info().Str("type", conditionName).Str("reason", condition.Reason).Msg("updated condition")
	}
	return changed
}

func formatMessageWithOptionalReason(baseFormat, reasonFormat, conditionMessage, reason string) string {
	if reason != "" {
		return fmt.Sprintf(reasonFormat, conditionMessage, reason)
	}
	return fmt.Sprintf(baseFormat, conditionMessage)
}
