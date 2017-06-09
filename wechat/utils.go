package wechat

// convertSexToGender 转换性别
func convertSexToGender(sex int) string {
	switch sex {
	case 1:
		return genderMale
	case 2:
		return genderFemale
	default:
		return genderOther
	}
}
