package ical

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
	"efs-ical/internal/efsapi"
)

// Build renders a VCALENDAR for one location's upcoming events.
func Build(calName string, events []efsapi.Event) *ics.Calendar {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//efs-ical//EFS Blood Donation Calendar//FR")
	cal.SetXWRCalName("Don du sang - " + calName)
	cal.SetXWRTimezone("Europe/Paris")

	now := time.Now().UTC()

	for _, ev := range events {
		uid := fmt.Sprintf("efs-%s@efs-ical.local", ev.RdvID)
		vev := cal.AddEvent(uid)
		vev.SetDtStampTime(now)
		vev.SetStartAt(ev.Start)
		vev.SetEndAt(ev.End)
		vev.SetSummary("Don du sang - " + ev.Title)
		vev.SetLocation(ev.Address)
		if ev.Lat != 0 || ev.Lon != 0 {
			vev.SetGeo(ev.Lat, ev.Lon)
		}
		if ev.URL != "" {
			vev.SetURL(ev.URL)
			vev.SetDescription("Prendre rendez-vous : " + ev.URL)
		}
	}

	return cal
}
