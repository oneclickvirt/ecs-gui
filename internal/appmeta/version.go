package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.188"
	UpstreamECSVersion = "v0.1.184"
)

func ReleaseVersion() string {
	return "v" + Version
}
