package report

import (
	"github.com/christoph-karpowicz/unifier/internal/server/synch"
	"github.com/christoph-karpowicz/unifier/internal/server/unifier"
)

type Reporter struct {
	synch *synch.Synch
	rep   *report
}

// CreateReporter creates a Reporter instance.
func CreateReporter(s *synch.Synch) unifier.Reporter {
	var newReporter Reporter = Reporter{
		s,
		CreateReport(s),
	}

	return &newReporter
}

// Init fills the necessary fields after the Synch instance finished its Init execution.
func (r *Reporter) Init() {
	for _, lnk := range r.synch.Links {
		_, lnkExists := r.rep.links[lnk.GetID()]
		if !lnkExists {
			r.rep.links[lnk.GetID()] = &link{Cmd: lnk.Cmd}
		}
	}
}

// AddAction adds a single action to the report.
// Action types:
// 	1.	idle - means two records that have been paired, but no action will be carried out because the relevant data is the same.
// 	2.	insert
// 	3. 	update
func (r *Reporter) AddAction(p unifier.Pairable, actionType string) (bool, error) {
	actionJSON, err := p.ReportJSON(actionType)
	if err != nil {
		return false, &SynchReportError{SynchName: r.synch.GetConfig().Name, ErrMsg: err.Error()}
	}

	switch actionType {
	case "idle":
		r.rep.links[p.GetLinkID()].Idle = append(r.rep.links[p.GetLinkID()].Idle, string(actionJSON))
	case "insert":
		r.rep.links[p.GetLinkID()].Inserts = append(r.rep.links[p.GetLinkID()].Inserts, string(actionJSON))
	case "update":
		r.rep.links[p.GetLinkID()].Updates = append(r.rep.links[p.GetLinkID()].Updates, string(actionJSON))
	}

	return true, nil
}

// Finalize wraps up the report creation process.
func (r *Reporter) Finalize() ([]byte, error) {
	if r.synch.IsSimulation() {
		r.rep.msg = "'" + r.synch.GetConfig().Name + "' simulation was successful. " +
			"The report contains changes that would be made if you requested an actual synchronization."
	} else {
		r.rep.msg = "'" + r.synch.GetConfig().Name + "' synchronization was successful. " +
			"The report contains changes that have been made to the relevant nodes."
	}

	toJSON, err := r.rep.ToJSON()
	if err != nil {
		return nil, &SynchReportError{SynchName: r.synch.GetConfig().Name, ErrMsg: err.Error()}
	}

	return toJSON, nil
}