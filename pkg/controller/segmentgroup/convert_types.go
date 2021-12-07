package segment

import (
	"github.com/haarchri/zpa-go-client/pkg/client/segment_group_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest SegmentGroup NotFound
	_, ok := err.(*segment_group_controller.GetSegmentGroupUsingGET1BadRequest)
	return ok
}
