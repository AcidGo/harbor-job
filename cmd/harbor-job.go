package main

import (
    "flag"
    "log"
    "regexp"
    "sort"

    "github.com/AcidGo/harbor-job/harbor"
    "github.com/AcidGo/harbor-job/rule"
    "gopkg.in/ini.v1"
)

const (
    CfgMainSec      = "main"
    cfgMainKeyUrl   = "harbor_url"
    cfgMainKeyVer   = "harbor_version"
    cfgMainKeyUser  = "harbor_user"
    cfgMainKeyPwd   = "harbor_pwd"
)

var (
    cfgPath     string
    harborUrl   string
    harborVer   string
    harborUser  string
    harborPwd   string
    rules       []*rule.Rule
)

func init() {
    flag.StringVar(&cfgPath, "f", "harbor-job.ini", "app main config file path")
    flag.Parse()

    cfg, err := ini.Load(cfgPath)
    if err != nil {
        log.Fatal(err)
    }

    mainSec, err := cfg.GetSection(CfgMainSec)
    if err != nil {
        log.Fatal(err)
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
            log.Printf("create rule from sec %s failed: %s\n", sec.Name, err.Error())
            continue
        }

        rules = append(rules, rule)
    }
}

func main() {
    client, err := harbor.NewClient(harborUrl, harborVer)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Login(harborUser, harborPwd)
    if err != nil {
        log.Fatal(err)
    }

    projects, err := client.Projects()
    if err != nil {
        log.Fatal(err)
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
            }
        }
        if projectId == 0 {
            log.Printf("not catch proejct %s, ignore it", rule.Project)
            continue
        }

        repos, err := client.Repositories(projectId)
        if err != nil {
            log.Fatal(err)
        }

        for _, repo := range repos {
            if !regexpMean(repo.Name, rule.RepoRegexp) {
                continue
            }
            log.Printf("%s mean the repo %s\n", rule.RepoRegexp, repo.Name)

            tags, err := client.Tags(repo.Name)
            if err != nil {
                log.Fatal(err)
            }

            sort.Sort(tags)
            if len(tags) < rule.RepoSaveNum {
                continue
            }

            for _, tag := range tags {
                log.Printf("tag's name: %s, tag's createtime: %s\n", tag.Name, tag.CreatedTime.Format("2006-01-02 15:04:05"))
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
            log.Printf("delIdx is %d, length of tags is %d", delIdx, len(tags))
            log.Printf("length of tagsSave: %d", len(tagsSave))

            if delIdx >= len(tags) {
                break
            }

            for _, tag := range tags[delIdx:] {
                if _, ok := tagsSave[tag.Digest]; ok {
                    continue
                }
                log.Printf("starting del tag: %s/%s", repo.Name, tag.Name)
                err := client.DeleteTag(repo.Name, tag.Name)
                if err != nil {
                    log.Printf("del tag %s/%s is failed", repo.Name, tag.Name)
                }
            }
        }

        projectDone[rule.Project] = projectId
        log.Printf("project %s done", rule.Project)
    }

    log.Println("all done")
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
