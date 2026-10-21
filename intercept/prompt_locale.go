package intercept

import "github.com/Philocyber/skopos/locale"

func init() {
	locale.Register(JudgeContextBoundary, JudgeContextBoundary)
	locale.Register(JudgeOutputContract, JudgeOutputContract)
	locale.Register(DefaultJudgePrompt, DefaultJudgePrompt)
}
