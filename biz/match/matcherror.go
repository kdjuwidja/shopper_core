package bizmatch

const (
	MatchItemIdNotProvided    = "match_item_id_not_provided"
	MatchFailedToProcess      = "match_failed_to_process"
	MatchShoplistItemNotFound = "match_shoplist_item_not_found"
)

type MatchError struct {
	Code    string
	Message string
}

func (e *MatchError) Error() string {
	return e.Message
}

func NewMatchError(code string, message string) *MatchError {
	return &MatchError{
		Code:    code,
		Message: message,
	}
}

func (e *MatchError) Is(target error) bool {
	return e.Code == target.(*MatchError).Code
}
