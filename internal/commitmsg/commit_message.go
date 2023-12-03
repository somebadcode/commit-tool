package commitmsg

type CommitMessage struct {
	Type     string
	Scope    string
	Subject  string
	Body     string
	Trailers map[string][]string
	Breaking bool
	Revert   bool
	Merge    bool
}
