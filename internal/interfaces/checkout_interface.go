package interfaces

import "trainiverse-backend/models"

type CheckoutInterface interface {
	SaveCheckout(checkoutData models.CheckoutData) error
}
