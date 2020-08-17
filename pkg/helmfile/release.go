package helmfile

type Release struct {
	Name             string
	Namespace        string
	Chart            string
	Version          string
	Values           []interface{}
	WorkingDirectory string
	Verify           bool
	Wait             bool
	Force            bool
	Atomic           bool
	CleanupOnFail    bool
	Timeout          int
	Kubeconfig       string
	Kubecontext      string
	Bin              string
	HelmBin          string
	DiffOutput       string
	ApplyOutput      string
}

func NewRelease(d ResourceRead) *Release {
	f := Release{}
	f.Namespace = d.Get(KeyNamespace).(string)
	f.Name = d.Get(KeyName).(string)
	if f.Name == "" {
		f.Name = d.Id()
	}
	f.Chart = d.Get(KeyChart).(string)
	f.Version = d.Get(KeyVersion).(string)
	f.Values = d.Get(KeyValues).([]interface{})
	f.WorkingDirectory = d.Get(KeyWorkingDirectory).(string)
	f.Verify = d.Get(KeyVerify).(bool)
	f.Wait = d.Get(KeyWait).(bool)
	f.Force = d.Get(KeyForce).(bool)
	f.Atomic = d.Get(KeyAtomic).(bool)
	f.CleanupOnFail = d.Get(KeyCleanupOnFail).(bool)
	f.Timeout = d.Get(KeyTimeout).(int)
	f.Kubeconfig = d.Get(KeyKubeconfig).(string)
	f.Kubecontext = d.Get(KeyKubecontext).(string)
	f.Bin = d.Get(KeyBin).(string)
	f.HelmBin = d.Get(KeyHelmBin).(string)
	f.DiffOutput = d.Get(KeyDiffOutput).(string)
	f.ApplyOutput = d.Get(KeyApplyOutput).(string)
	return &f
}
