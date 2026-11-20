package main

type Manifest struct {
	Version   int      `json:"version"`
	CreatedAt string   `json:"created_at"`
	Host      HostInfo `json:"host"`
	Git       GitInfo  `json:"git"`
	Command   CmdInfo  `json:"command"`
	Files     []string `json:"files"`
	Note      string   `json:"note,omitempty"`
}

type HostInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
}

type GitInfo struct {
	SHA    string `json:"sha"`
	Branch string `json:"branch"`
	Remote string `json:"remote"`
	Dirty  bool   `json:"dirty"`
}

type CmdInfo struct {
	Argv       []string `json:"argv,omitempty"`
	ExitCode   *int     `json:"exit_code,omitempty"`
	DurationMS int64    `json:"duration_ms,omitempty"`
	Cwd        string   `json:"cwd"`
}
