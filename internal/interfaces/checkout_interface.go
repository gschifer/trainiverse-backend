package interfaces

import "trainiverse-backend/internal/models"

type CheckoutInterface interface {
	SaveCheckout(checkoutData models.CheckoutData) error
}
