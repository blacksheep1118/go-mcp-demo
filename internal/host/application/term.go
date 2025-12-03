package application

import (
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/west2-online/jwch"
)

func (h *Host) GetTermList() (*jwch.SchoolCalendar, error) {
	calendar, err := jwch.NewStudent().GetSchoolCalendar()
	if err = base.HandleJwchError(err); err != nil {
		return nil, fmt.Errorf("service.GetTermList: Get term list failed %w", err)
	}
	return calendar, nil
}

func (h *Host) GetTerm(req *api.TermRequest) (bool, *jwch.CalTermEvents, error) {
	var err error
	var events *jwch.CalTermEvents

	events, err = jwch.NewStudent().GetTermEvents(req.Term)
	if err = base.HandleJwchError(err); err != nil {
		return false, nil, fmt.Errorf("service.GetTerm: Get term  failed %w", err)
	}
	return true, events, err
}
