package bizshoplist

const (
	ShoplistNotFound        = "shoplist_not_found"
	ShoplistNotOwned        = "shoplist_not_owned"
	ShoplistNotMember       = "shoplist_not_member"
	ShoplistNotOwner        = "shoplist_not_owner"
	ShoplistFailedToCreate  = "shoplist_failed_to_create"
	ShoplistFailedToProcess = "shoplist_failed_to_process"
	ShoplistFailedToUpdate  = "shoplist_failed_to_update"
	ShoplistItemNameEmpty   = "shoplist_item_name_empty"
	ShoplistItemNotFound    = "shoplist_item_not_found"
)

type ShoplistError struct {
	ErrCode string
	Message string
}

func (e *ShoplistError) Error() string {
	return e.Message
}

func NewShoplistError(code string, message string) *ShoplistError {
	return &ShoplistError{
		ErrCode: code,
		Message: message,
	}
}

func (e *ShoplistError) Is(target error) bool {
	return e.ErrCode == target.(*ShoplistError).ErrCode
}
