package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "time"

    "github.com/AcidGo/harbor-job/harbor"
    "github.com/AcidGo/harbor-job/logger"
    "github.com/AcidGo/harbor-job/rule"
    "gopkg.in/ini.v1"
)

const (
    CfgMainSec          = "main"
    cfgMainKeyUrl       = "harbor_url"
    cfgMainKeyVer       = "harbor_version"
    cfgMainKeyUser      = "harbor_user"
    cfgMainKeyPwd       = "harbor_pwd"
    cfgMainKeyLogDir    = "log_dir"
    cfgMainKeyLogName   = "log_name"
    cfgMainKeyLogLevel  = "log_level"
    cfgMainKeyLogReport = "log_report"
)

var (
    // param
    cfgPath     string
    dryRun      bool
    meanRepo    bool
    
    // config
    harborUrl   string
    harborVer   string
    harborUser  string
    harborPwd   string
    logDir      string
    logName     string
    logLevel    string
    logReport   bool

    // logger
    logging     *logger.ContextLogger

    // runtime
    rules       []*rule.Rule

    // app info
    AppName             string
    AppAuthor           string
    AppVersion          string
    AppGitCommitHash    string
    AppBuildTime        string
    AppGoVersion        string
)

func init() {
    logging = logger.FitContext("harbor-job")

    flag.StringVar(&cfgPath, "f", "harbor-job.ini", "app main config file path")
    flag.BoolVar(&dryRun, "dry-run", false, "execute command with dry run mode")
    flag.BoolVar(&meanRepo, "mean-repo", false, "show meant repo name")
    flag.Usage = flagUsage
    flag.Parse()

    cfg, err := ini.Load(cfgPath)
    if err != nil {
        logging.Fatal(err)
    }

    mainSec, err := cfg.GetSection(CfgMainSec)
    if err != nil {
        logging.Fatal(err)
    }

    // init logger
    logDir          = mainSec.Key(cfgMainKeyLogDir).String()
    logName         = mainSec.Key(cfgMainKeyLogName).String()
    logLevel        = mainSec.Key(cfgMainKeyLogLevel).String()
    logReport, err   = mainSec.Key(cfgMainKeyLogReport).Bool()
    if err != nil {
        logging.Fatal(err)
    }

    if !IsDir(logDir) {
        logging.Fatalf("the log dir %s is not a dir or no exists", logDir)
    }

    logPath := filepath.Join(logDir, logName)
    err = logger.LogFileSetting(logPath)
    if err != nil {
        logging.Fatal(err)
    }
    logger.ReportCallerSetting(logReport)
    err = logger.LogLevelSetting(logLevel)
    if err != nil {
        logging.Fatal(err)
    }

    harborUrl = mainSec.Key(cfgMainKeyUrl).String()
    harborVer = mainSec.Key(cfgMainKeyVer).String()
    harborUser = mainSec.Key(cfgMainKeyUser).String()
    harborPwd = mainSec.Key(cfgMainKeyPwd).String()

    rules = make([]*rule.Rule, 0)

    for _, sec := range cfg.Sections() {
        if sec.Name() == ini.DEFAULT_SECTION || sec.Name() == CfgMainSec {
            continue
        }

        rule := &rule.Rule{
            Project: sec.Name(),
        }
        err := sec.MapTo(rule)
        if err != nil {
            logging.Errorf("create rule from sec %s failed: %s", sec.Name(), err.Error())
            continue
        }

        rules = append(rules, rule)
    }
}

func main() {
    client, err := harbor.NewClient(harborUrl, harborVer)
    if err != nil {
        logging.Fatal(err)
    }

    err = client.Login(harborUser, harborPwd)
    if err != nil {
        logging.Fatal(err)
    }

    projects, err := client.Projects()
    if err != nil {
        logging.Fatal(err)
    }

    projectDone := make(map[string]int)

    for _, rule := range rules {
        if _, ok := projectDone[rule.Project]; ok {
            continue
        }

        projectId := 0
        for _, p := range projects {
            if p.Name == rule.Project {
                projectId = p.ProjectId
                break
            }
        }
        if projectId == 0 {
            logging.Warnf("not catch proejct %s, ignore it", rule.Project)
            continue
        }

        repos, err := client.Repositories(projectId)
        if err != nil {
            logging.Fatal(err)
        }
        logging.Infof("found %d repositories in the project %s", len(repos), rule.Project)
        logging.Tracef("%#v", repos)

        for _, repo := range repos {
            logging.Debugf("start deal with repo %s", repo.Name)

            if !regexpMean(repo.Name, rule.RepoRegexp) {
                logging.Debugf("repo %s not mean the regexp rule, ignore it", repo.Name)
                continue
            }
            logging.Debugf("%s mean the repo %s", rule.RepoRegexp, repo.Name)

            if meanRepo {
                continue
            }

            tags, err := client.Tags(repo.Name)
            if err != nil {
                logging.Fatal(err)
            }
            logging.Debugf("found %d tags in the repo %s", len(tags), repo.Name)

            sort.Sort(tags)
            if len(tags) < rule.RepoSaveNum {
                continue
            }

            for _, tag := range tags {
                logging.Debugf("tag's name: %s, tag's createtime: %s", tag.Name, tag.CreatedTime.Format("2006-01-02 15:04:05"))
            }

            tagsSave := make(map[string]string)
            delIdx := 0
            for idx, tag := range tags {
                if len(tagsSave) >= rule.RepoSaveNum {
                    break
                }

                if _, ok := tagsSave[tag.Digest]; ok {
                    continue
                }

                tagsSave[tag.Digest] = tag.Name
                delIdx = idx + 1
            }
            logging.Infof("delIdx is %d, length of tags is %d", delIdx, len(tags))
            logging.Infof("length of tagsSave: %d", len(tagsSave))

            if delIdx >= len(tags) {
                continue
            }

            for _, tag := range tags[delIdx:] {
                if _, ok := tagsSave[tag.Digest]; ok {
                    continue
                }
                logging.Infof("starting del tag: %s/%s", repo.Name, tag.Name)
                if !dryRun {
                    err := client.DeleteTag(repo.Name, tag.Name)
                    if err != nil {
                        logging.Infof("del tag %s/%s is failed", repo.Name, tag.Name)
                    }
                    time.Sleep(200*time.Millisecond)
                }
            }
        }

        projectDone[rule.Project] = projectId
        logging.Infof("project %s done", rule.Project)
    }

    logging.Info("all done")
}

func regexpMean(src, re string) (bool) {
    cp, err := regexp.Compile(re)
    if err != nil {
        if src == re {
            return true
        }
        return false
    }

    return cp.MatchString(src)
}

func flagUsage() {
    usageMsg := fmt.Sprintf(`App: %s
Version: %s
Author: %s
GitCommit: %s
BuildTime: %s
GoVersion: %s
Options:
`, AppName, AppVersion, AppAuthor, AppGitCommitHash, AppBuildTime, AppGoVersion)

    fmt.Fprintf(os.Stderr, usageMsg)
    flag.PrintDefaults()
}

func IsDir(path string) (bool) {
    s, err := os.Stat(path)
    if err != nil {
        return false
    }
    return s.IsDir()
}