package server

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/Philocyber/skopos/locale"
)

func init() {
	locale.Register(reporterToolCallMessage, reporterToolCallMessage, "방금 report_finding으로 취약점이 등록되었습니다. 반환 JSON에서 finding_id(독립 취약점 레코드)와 finding_node_id(탐색 노드)를 읽으세요. get_finding_traffic(finding_id)로 증거 목록과 version을 읽으세요. 빈 목록도 정상이며 보고서 작성을 막지 않습니다. 실행 지침에서 자동 연결이 켜져 있으면 읽기 전에 이 취약점의 트래픽을 검증/연결하세요. 노드 상세에는 finding_node_id를 사용하세요. 마지막으로 update_finding_report(finding_id=finding_node_id, report, evidence_version=실제로 읽은 버전)로 저장하세요. evidence_version을 전달하지 않으면 보고서가 계속 오래된 상태로 표시됩니다. 두 ID 체계를 혼용하지 마세요.")
}

// reporterTriggerText localizes only exact bundled trigger messages, including
// their historical stock versions. Operator-authored trigger content is preserved.
func reporterTriggerText(text string, lang locale.Lang) string {
	digest := sha256.Sum256([]byte(text))
	if text == reporterToolCallMessage || text == locale.Text(locale.Ko, reporterToolCallMessage) || isLegacyReporterTrigger(text) || hex.EncodeToString(digest[:]) == "2b766affdba89c92772698367e41fc6d570ee25f29964576c0aec96da4d176e6" {
		return locale.Text(lang, reporterToolCallMessage)
	}
	return text
}
