package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.174"
	UpstreamECSVersion = "v0.1.171"
)

func ReleaseVersion() string {
	return "v" + Version
}
