package apiHandlers

import "net/http"

const (
	ErrInvalidToken                           = "GEN_00001"
	ErrInvalidRequestBody                     = "GEN_00002"
	ErrMissingRequiredField                   = "GEN_00003"
	ErrMissingRequiredParam                   = "GEN_00004"
	ErrInternalServerError                    = "GEN_99999"
	ErrInvalidPostalCode                      = "USR_00001"
	ErrUserProfileNotFound                    = "USR_00002"
	ErrShoplistNotFound                       = "SHP_00001"
	ErrShoplistNotOwned                       = "SHP_00002"
	ErrShoplistItemNotFound                   = "SHP_00003"
	ErrMissingRequiredFieldUpdateShoplistItem = "SHP_00004"
)

var responseMap = map[string]response{
	ErrInternalServerError:                    {ErrInternalServerError, http.StatusInternalServerError, "Internal server error."},
	ErrInvalidToken:                           {ErrInvalidToken, http.StatusUnauthorized, "Invalid or missing bearer token."},
	ErrInvalidRequestBody:                     {ErrInvalidRequestBody, http.StatusBadRequest, "Invalid request body"},
	ErrMissingRequiredField:                   {ErrMissingRequiredField, http.StatusBadRequest, "Missing field in body: %s"},
	ErrMissingRequiredParam:                   {ErrMissingRequiredParam, http.StatusBadRequest, "Missing parameter: %s"},
	ErrInvalidPostalCode:                      {ErrInvalidPostalCode, http.StatusBadRequest, "Invalid postal code."},
	ErrUserProfileNotFound:                    {ErrUserProfileNotFound, http.StatusNotFound, "User profile not found."},
	ErrShoplistNotFound:                       {ErrShoplistNotFound, http.StatusNotFound, "Shoplist not found."},
	ErrShoplistNotOwned:                       {ErrShoplistNotOwned, http.StatusForbidden, "Only the owner can perform this action."},
	ErrShoplistItemNotFound:                   {ErrShoplistItemNotFound, http.StatusNotFound, "Item not found."},
	ErrMissingRequiredFieldUpdateShoplistItem: {ErrMissingRequiredFieldUpdateShoplistItem, http.StatusBadRequest, "Request body must include at least one of item_name, brand_name, extra_info or Is_bought."},
}
