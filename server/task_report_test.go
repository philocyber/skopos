package server

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Philocyber/skopos/db"
)

func TestTaskReportUsesTaskScopeAndStandaloneFindings(t *testing.T) {
	s, recorded, request := trafficEvidenceServer(t)
	finding, err := s.m.pg.GetFinding(recorded.FindingID)
	if err != nil || finding.TaskID == nil {
		t.Fatalf("fixture finding task: %v", err)
	}
	tid := *finding.TaskID
	task, ok := s.m.Task(strconv.FormatInt(tid, 10))
	if !ok {
		t.Fatal("fixture task is missing")
	}
	marker := strconv.FormatInt(time.Now().UnixNano(), 10)
	inDomain, outDomain := "in-"+marker+".invalid", "out-"+marker+".invalid"
	inAsset, err := s.m.Assets().UpsertRootDomain(db.UpsertRootDomainReq{Domain: inDomain, TaskID: tid})
	if err != nil {
		t.Fatal(err)
	}
	outAsset, err := s.m.Assets().UpsertRootDomain(db.UpsertRootDomainReq{Domain: outDomain})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.m.Assets().DeleteByIDs([]int64{inAsset, outAsset}) })
	if _, err := s.m.Assets().AddAgentScope(tid, "root_domain", inDomain, "Synthetic authorized scope", "manual"); err != nil {
		t.Fatal(err)
	}
	if _, err := task.Store.AddNode(db.KindFact, map[string]any{"summary": "SYNTHETIC_NEGATIVE_OBSERVATION"}, 0, "confirmed", "test", []int64{inAsset}); err != nil {
		t.Fatal(err)
	}
	if _, err := task.Store.AddNode(db.KindIntent, map[string]any{"summary": "SYNTHETIC_INCOMPLETE_ATTEMPT"}, 0, "blocked", "test", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.m.pg.AddFinding(tid, 0, "TEST", "SYNTHETIC_STANDALONE_FINDING", "high", "Needs validation", "Synthetic evidence", "test", nil); err != nil {
		t.Fatal(err)
	}
	w := request("GET", fmt.Sprintf("/api/report?task=%d", tid), "")
	if w.Code != 200 {
		t.Fatalf("report: %d %s", w.Code, w.Body.String())
	}
	text := w.Body.String()
	for _, want := range []string{inDomain, "Registered findings: **2**", "SYNTHETIC_STANDALONE_FINDING", "SYNTHETIC_NEGATIVE_OBSERVATION", "SYNTHETIC_INCOMPLETE_ATTEMPT", "Review status: pending", "Assets with recorded results: **1 / 1**"} {
		if !strings.Contains(text, want) {
			t.Errorf("report missing %q", want)
		}
	}
	if strings.Contains(text, outDomain) {
		t.Fatal("unrelated shared asset leaked into task report")
	}
}
