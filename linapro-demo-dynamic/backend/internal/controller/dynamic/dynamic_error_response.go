// This file translates dynamic demo business errors into bridge responses
// via the shared pluginbridge ErrorClassifier composition.

package dynamic

import (
	dynamicservice "lina-plugin-linapro-demo-dynamic/backend/internal/service/dynamic"

	bridgeguest "lina-core/pkg/plugin/pluginbridge/guest"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// dynamicErrorClassifier maps dynamic sample business errors to normalized
// bridge responses. BindJSON sentinels are handled by pluginbridge itself.
var dynamicErrorClassifier = bridgeguest.NewErrorClassifier(
	bridgeguest.NewErrorCase(dynamicservice.IsDemoRecordInvalidInput, protocol.NewBadRequestResponse),
	bridgeguest.NewErrorCase(dynamicservice.IsDemoRecordNotFound, protocol.NewNotFoundResponse),
)

// wrapDynamicError converts one dynamic sample business error into a
// prebuilt bridge response error so typed guest controllers can return it
// through the standard error channel.
func wrapDynamicError(err error) error {
	if err == nil {
		return nil
	}
	return bridgeguest.NewResponseError(dynamicErrorClassifier.Classify(err))
}
