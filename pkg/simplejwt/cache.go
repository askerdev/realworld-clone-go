package simplejwt

import "time"

type Cache interface {
	Get(key uint64) (time.Time, bool)
	Set(key uint64, value time.Time) error
}
