package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.149"
	UpstreamECSVersion = "v0.1.149"
)

func ReleaseVersion() string {
	return "v" + Version
}
