package domain

// ReservationStatus represents the status of a stock reservation
type ReservationStatus string

const (
	// ReservationStatusPending indicates the reservation is active and stock is held
	ReservationStatusPending ReservationStatus = "pending"

	// ReservationStatusFulfilled indicates the reservation was fulfilled (stock deducted)
	ReservationStatusFulfilled ReservationStatus = "fulfilled"

	// ReservationStatusCancelled indicates the reservation was cancelled (stock released)
	ReservationStatusCancelled ReservationStatus = "cancelled"

	// ReservationStatusExpired indicates the reservation expired (stock released)
	ReservationStatusExpired ReservationStatus = "expired"
)

// IsValid checks if the reservation status is valid
func (s ReservationStatus) IsValid() bool {
	switch s {
	case ReservationStatusPending,
		ReservationStatusFulfilled,
		ReservationStatusCancelled,
		ReservationStatusExpired:
		return true
	default:
		return false
	}
}

// IsFinal checks if the reservation status is a final state
func (s ReservationStatus) IsFinal() bool {
	return s == ReservationStatusFulfilled ||
		s == ReservationStatusCancelled ||
		s == ReservationStatusExpired
}
