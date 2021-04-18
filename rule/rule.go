package rule

type Rule struct {
    Project     string
    RepoRegexp  string  `ini:"repo_regexp"`
    RepoSaveNum int     `ini:"repo_save_num"`
}

type Rules []*Rule