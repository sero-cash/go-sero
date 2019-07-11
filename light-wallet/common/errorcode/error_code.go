package errorcode

const SUCCESS_CODE  = "SUCCESS"
const SUCCESS_DESC  = "Operation is successful"

const FAIL_CODE  = "FAIL"
const FAIL_DESC  = "The operation failure"


//系统内部异常
const SystemInternalError = "F0001"
//token鉴权失败
const AuthenticationFailed = "F0002"
//验证码无效
const InvalidVerificationCode = "F0003"
//用户名或密码不对
const IncorrectUserOrPassword = "F0004"
//业务参数校验失败
const InvalidBizParameters = "F0005"

//基本参数校验失败
const InvalidBaseParameters = "F0006"

//无效的Token
const InvalidToken = "F0007"
//支付密码不对
const InvalidPayPassword = "F0008"
//用户没有找到
const UserNotExist = "F0009"
//用户没有绑定手机
const NotBandingMobile = "F0010"
//无效的图片验证码
const InvalidCaptcha = "F0011"
