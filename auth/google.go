package auth

import (
	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
	"context"
	"fmt"
	"login/exception"
)

func sendToGoogle(token string, recaptchaAction string) bool {
	// 待办事项：在运行示例之前，替换令牌和 reCAPTCHA 操作变量。
	projectID := "6Lexl4sqAAAAAKNXoHp_PuPCbl-y9t6ZkD2tFI4C"
	recaptchaKey := "6Lexl4sqAAAAAOzkLKgxOgrg5dj7gu1_mKc51N6w"
	return createAssessment(projectID, recaptchaKey, token, recaptchaAction)
}

/**
 * 创建评估以分析界面操作的风险。
 *
 * @param projectID: 您的 Google Cloud 项目 ID。
 * @param recaptchaKey: 与网站/应用关联的 reCAPTCHA 密钥
 * @param token: 从客户端获取的已生成令牌。
 * @param recaptchaAction: 与令牌对应的操作名称。
 */
func createAssessment(projectID string, recaptchaKey string, token string, recaptchaAction string) bool {

	exception.PrintWarning(createAssessment, fmt.Errorf("当前capcha服务不可用，因此这里比较卡并且有警告，请不用大惊小怪"))

	// 创建 reCAPTCHA 客户端。
	ctx := context.Background()
	client, err := recaptcha.NewClient(ctx)
	if err != nil {
		exception.PrintWarning(createAssessment, err)
		return true
	}

	// 设置要跟踪的事件的属性。
	event := &recaptchapb.Event{
		Token:   token,
		SiteKey: recaptchaKey,
	}

	assessment := &recaptchapb.Assessment{
		Event: event,
	}

	// 构建评估请求。
	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: assessment,
		Parent:     fmt.Sprintf("projects/%s", projectID),
	}

	response, err := client.CreateAssessment(
		ctx,
		request)

	if err != nil {
		exception.PrintWarning(createAssessment, err)
		return true
	}

	if response.RiskAnalysis.Score > 0.5 {
		fmt.Printf("The reCAPTCHA score for this token is:  %v", response.RiskAnalysis.Score)
		for _, reason := range response.RiskAnalysis.Reasons {
			fmt.Printf(reason.String() + "\n")
		}
		return true
	} else {
		fmt.Printf("The reCAPTCHA score for this token is:  %v", response.RiskAnalysis.Score)
		for _, reason := range response.RiskAnalysis.Reasons {
			fmt.Printf(reason.String() + "\n")
		}
		return false
	}

	//// 检查令牌是否有效。
	//if !response.TokenProperties.Valid {
	//	fmt.Printf("The CreateAssessment() call failed because the token was invalid for the following reasons: %v",
	//		response.TokenProperties.InvalidReason)
	//	return
	//}
	//
	//// 检查是否执行了预期操作。
	//if response.TokenProperties.Action != recaptchaAction {
	//	fmt.Printf("The action attribute in your reCAPTCHA tag does not match the action you are expecting to score")
	//	return
	//}

	//// 获取风险得分和原因。
	//// 如需详细了解如何解读评估，请参阅：
	//// https://cloud.google.com/recaptcha-enterprise/docs/interpret-assessment
	//fmt.Printf("The reCAPTCHA score for this token is:  %v", response.RiskAnalysis.Score)
	//
	//for _, reason := range response.RiskAnalysis.Reasons {
	//	fmt.Printf(reason.String() + "\n")
	//}
}
