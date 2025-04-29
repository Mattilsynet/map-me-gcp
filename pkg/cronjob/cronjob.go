package cronjob

import (
	"github.com/Mattilsynet/map-me-gcp/gen/mattilsynet/cronjob/cronjob"
)
func RegisterCronHandler(fn func()) {
   cronjob.Exports.CronHandler = fn
}
