package segment

import (
	"errors"

	"github.com/haarchri/zpa-go-client/pkg/client/segment_group_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest SegmentGroup NotFound
	var badRequest *segment_group_controller.GetSegmentGroupUsingGET1BadRequest
	return errors.As(err, &badRequest)
}
